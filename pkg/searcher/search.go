package searcher

import (
	"bufio"
	"io/fs"
	"sync"

	"word-search-in-files/pkg/internal/dir"
)

type WordSearcher interface {
	Search(word string) []string
	ConstructFileDictionary() error
}

func NewSearcher(fs fs.FS, dir string) *Searcher {
	return &Searcher{
		FS:  fs,
		Dir: dir,
	}
}

type Searcher struct {
	FS             fs.FS
	Dir            string
	FileDictionary map[string]map[string]struct{}
}

func (s *Searcher) ConstructFileDictionary() error {
	s.FileDictionary = make(map[string]map[string]struct{})
	mu := &sync.Mutex{}
	fileNames, err := dir.FilesFS(s.FS, s.Dir)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	errChan := make(chan error)
	wg.Add(len(fileNames))
	for _, fName := range fileNames {
		s.FileDictionary[fName] = make(map[string]struct{})
		go func(fName string) {
			var errorGoroutine error
			defer wg.Done()
			file, errorGoroutine := s.FS.Open(fName)
			if errorGoroutine != nil {
				errChan <- errorGoroutine
				return
			}
			defer func() {
				err = file.Close()
				if err != nil {
					return
				}
			}()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				currentWord := make([]byte, 0)
				for i := 0; i < len(line); i++ {
					if line[i] != ' ' {
						currentWord = append(currentWord, line[i])
						continue
					}
					if len(currentWord) != 0 {
						mu.Lock()
						s.FileDictionary[fName][string(currentWord)] = struct{}{}
						mu.Unlock()
						currentWord = currentWord[:0]
					}
				}
				if len(currentWord) != 0 {
					mu.Lock()
					s.FileDictionary[fName][string(currentWord)] = struct{}{}
					mu.Unlock()
				}
			}
			if errorGoroutine = scanner.Err(); errorGoroutine != nil {
				errChan <- errorGoroutine
				return
			}

		}(fName)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err = range errChan {
		return err
	}
	return nil

}

func (s *Searcher) Search(word string) []string {
	files := make([]string, 0)
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for fName := range s.FileDictionary {
		wg.Add(1)
		go func(fName string) {
			defer wg.Done()
			if _, exists := s.FileDictionary[fName][word]; exists {
				mu.Lock()
				files = append(files, fName)
				mu.Unlock()
			}
		}(fName)
	}
	wg.Wait()
	if len(files) == 0 {
		return nil
	}
	return files
}
