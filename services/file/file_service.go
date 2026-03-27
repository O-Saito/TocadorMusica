package file

import (
	"os"
	"path/filepath"
	"strings"

	"tocadormusica/domain"
)

var audioExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".ogg":  true,
	".m4a":  true,
	".aac":  true,
	".wma":  true,
}

var videoExtensions = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mkv":  true,
	".mov":  true,
	".webm": true,
}

type FileService struct{}

func New() domain.FileService {
	return &FileService{}
}

func (s *FileService) isAudioOrVideo(ext string) bool {
	ext = strings.ToLower(ext)
	return audioExtensions[ext] || videoExtensions[ext]
}

func (s *FileService) collectFiles(dir string, recursive bool) []string {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if recursive {
				files = append(files, s.collectFiles(filepath.Join(dir, entry.Name()), true)...)
			}
			continue
		}

		ext := filepath.Ext(entry.Name())
		if s.isAudioOrVideo(ext) {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files
}

func (s *FileService) Search(folders []string, query string, recursive bool) ([]domain.Track, error) {
	if len(folders) == 0 || query == "" {
		return nil, nil
	}

	query = strings.ToLower(query)
	var results []domain.Track

	for _, folder := range folders {
		if folder == "" {
			continue
		}

		files := s.collectFiles(folder, recursive)

		for _, path := range files {
			name := strings.ToLower(filepath.Base(path))
			if !strings.Contains(name, query) {
				continue
			}

			track := domain.NewTrackFromFile(path)
			results = append(results, track)
		}
	}

	return results, nil
}

func (s *FileService) ListAll(folders []string, recursive bool) ([]domain.Track, error) {
	if len(folders) == 0 {
		return nil, nil
	}

	var results []domain.Track

	for _, folder := range folders {
		if folder == "" {
			continue
		}

		files := s.collectFiles(folder, recursive)

		for _, path := range files {
			track := domain.NewTrackFromFile(path)
			results = append(results, track)
		}
	}

	return results, nil
}
