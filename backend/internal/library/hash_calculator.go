package library

import (
	"github.com/upper/db/v4"
	"runtime"
	"sort"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/internal/database"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/duplo"
)

const maxSimilarImages = 20

type HashCalculator struct {
	stopChannel     chan bool
	outputChannel   chan *HashResult
	imageLoader     api.ImageLoader
	threadCount     int
	similarityIndex *database.SimilarityIndex
	hashIndex       *duplo.Store
}

func NewHashCalculator(similarityIndex *database.SimilarityIndex, imageLoader api.ImageLoader, threadCount int) *HashCalculator {
	return &HashCalculator{
		similarityIndex: similarityIndex,
		imageLoader:     imageLoader,
		hashIndex:       duplo.New(),
		threadCount:     threadCount,
	}
}

func (s *HashCalculator) GenerateHashes(images []*apitype.ImageFile, statusCallback func(int, int)) (map[apitype.ImageId]*duplo.Hash, error) {
	startTime := time.Now()
	hashExpected := len(images)
	logger.Info.Printf("Generate hashes for %d images...", hashExpected)
	statusCallback(0, hashExpected)
	hashes := map[apitype.ImageId]*duplo.Hash{}

	if hashExpected > 0 {
		// Just to make things consistent in case Go decides to change the default
		logger.Info.Printf(" * Using %d threads", s.threadCount)
		runtime.GOMAXPROCS(s.threadCount)

		s.stopChannel = make(chan bool)
		inputChannel := make(chan *apitype.ImageFile, hashExpected)
		s.outputChannel = make(chan *HashResult)

		// Add images to input queue for goroutines
		for _, imageFile := range images {
			inputChannel <- imageFile
		}

		// Spin up goroutines which will process the data
		// only same number as CPU cores so that we will only max X hashes are
		// processed at once. Otherwise the goroutines might start processing
		// all images at once which would use all available RAM
		for i := 0; i < s.threadCount; i++ {
			go hashImage(inputChannel, s.outputChannel, s.stopChannel, s.imageLoader)
		}

		var mux sync.Mutex

		var totalHashesProcessed = 0
		var successfulHashes = 0
		var errors []error
		for result := range s.outputChannel {
			totalHashesProcessed++

			if result.err == nil {

				s.addHashToMap(result, hashes, &mux)

				statusCallback(totalHashesProcessed, hashExpected)
				s.hashIndex.Add(result.imageId, *result.hash)
				successfulHashes++
			} else {
				errors = append(errors, result.err)
			}

			if totalHashesProcessed == hashExpected {
				s.StopHashes()
			}
		}
		close(inputChannel)

		endTime := time.Now()
		d := endTime.Sub(startTime)
		logger.Info.Printf("%d hashes generated in %s (%d errors)", hashExpected, d.String(), len(errors))

		if len(errors) > 0 {
			logger.Error.Printf("Errors while processing hashes")
			for _, err := range errors {
				logger.Error.Printf(" - %s", err)
			}
		}

		avg := d.Milliseconds() / int64(successfulHashes)
		// Remember to take thread count otherwise the avg time is too small
		f := time.Millisecond * time.Duration(avg) * time.Duration(s.threadCount)
		logger.Info.Printf("  On average: %s/image", f.String())

		// Always notify that everything processed
		statusCallback(hashExpected, hashExpected)
	} else {
		logger.Info.Printf("No hashes to generate")
	}

	return hashes, nil
}

func (s *HashCalculator) StopHashes() {
	if s.stopChannel != nil {
		for i := 0; i < s.threadCount; i++ {
			s.stopChannel <- true
		}
		close(s.outputChannel)
		close(s.stopChannel)
		s.stopChannel = nil
	}
}

func (s *HashCalculator) BuildSimilarityIndex(hashes map[apitype.ImageId]*duplo.Hash, statusCallback func(int, int)) error {
	logger.Info.Printf("Building similarity index for %d most similar images for each image", maxSimilarImages)

	startTime := time.Now()

	err := s.similarityIndex.DoInTransaction(func(session db.Session) error {
		if err := s.similarityIndex.StartRecreateSimilarImageIndex(session); err != nil {
			logger.Error.Print("Error while clearing similar images", err)
			return err
		}

		numOfHashedImages := len(hashes)
		statusCallback(0, len(hashes))
		imageIndex := 0
		for imageId, hash := range hashes {
			searchStart := time.Now()
			matches := s.hashIndex.Query(*hash)
			searchEnd := time.Now()

			sortStart := time.Now()
			sort.Sort(matches)
			sortEnd := time.Now()

			addStart := time.Now()
			i := 0
			for _, match := range matches {
				similarId := match.ID.(apitype.ImageId)
				if imageId != similarId {
					if err := s.similarityIndex.
						AddSimilarImage(imageId, similarId, i, match.Score); err != nil {
						logger.Error.Print("Error while storing similar images", err)
						return err
					}
					i++
				}
				if i == maxSimilarImages {
					break
				}
			}
			addEnd := time.Now()

			statusCallback(imageIndex, numOfHashedImages)
			imageIndex = imageIndex + 1

			logger.Trace.Printf("Print added matches for image: %d", imageId)
			logger.Trace.Printf(" -    Search: %s", searchEnd.Sub(searchStart))
			logger.Trace.Printf(" -      Sort: %s", sortEnd.Sub(sortStart))
			logger.Trace.Printf(" -       Add: %s", addEnd.Sub(addStart))
			logger.Trace.Printf(" - Add/image: %s", addEnd.Sub(addStart)/maxSimilarImages)
		}
		if err := s.similarityIndex.EndRecreateSimilarImageIndex(); err != nil {
			logger.Error.Print("Error while finishing similar images index", err)
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	endTime := time.Now()
	d := endTime.Sub(startTime)
	logger.Info.Printf("Similarity index has been built in %s", d.String())
	return nil
}

func (s *HashCalculator) addHashToMap(result *HashResult, hashes map[apitype.ImageId]*duplo.Hash, mux *sync.Mutex) {
	mux.Lock()
	defer mux.Unlock()
	hashes[result.imageId] = result.hash
}
