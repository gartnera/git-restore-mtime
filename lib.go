package main

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	ignore "github.com/crackcomm/go-gitignore"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ManagerOptionT func(*Manager)

// WithMaxDepth sets the maximum depth to traverse the commit history
func WithMaxDepth(depth int) ManagerOptionT {
	return func(m *Manager) {
		m.maxDepth = depth
	}
}

type Manager struct {
	modTimes      map[string]time.Time
	repo          *git.Repository
	repoRoot      string
	ignore        *ignore.GitIgnore
	minCommitTime time.Time
	maxDepth      int
}

func NewManagerFromPath(repoRoot string, opts ...ManagerOptionT) (*Manager, error) {
	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return nil, err
	}
	return newManager(r, repoRoot, opts...), nil
}

func newManager(repo *git.Repository, repoRoot string, opts ...ManagerOptionT) *Manager {
	i, err := ignore.CompileIgnoreFile(path.Join(repoRoot, ".gitignore"))
	if err != nil {
		slog.Warn("failed to read .gitignore", "err", err)
	}
	m := &Manager{
		modTimes:      make(map[string]time.Time),
		repo:          repo,
		repoRoot:      repoRoot,
		ignore:        i,
		minCommitTime: time.Now(),
		maxDepth:      math.MaxInt64,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) updateTimeIfGreater(path string, commitTime time.Time) {
	currentTime, ok := m.modTimes[path]
	if ok && currentTime.After(commitTime) {
		return
	}
	m.modTimes[path] = commitTime
}

func (m *Manager) handleDiffPath(path string, commitTime time.Time) {
	if path == "." {
		return
	}
	slog.Debug("handling diff path", "path", path, "commitTime", commitTime)
	m.updateTimeIfGreater(path, commitTime)
	// recurse up directory structure
	dirName := filepath.Dir(path)
	m.handleDiffPath(dirName, commitTime)
}

func (m *Manager) SetFromGit() error {
	head, err := m.repo.Head()
	if err != nil {
		return fmt.Errorf("get head: %w", err)
	}
	currentCommit, err := m.repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("get commit object: %w", err)
	}
	depth := 0
	for {
		currentTree, err := currentCommit.Tree()
		if err != nil {
			return fmt.Errorf("get current tree: %w", err)
		}
		currentCommitTime := currentCommit.Committer.When
		parentCommit, err := currentCommit.Parent(0)
		if err != nil {
			if err != object.ErrParentNotFound {
				return fmt.Errorf("get parent commit: %w", err)
			}
			slog.Debug("got root commit", "currentCommit", currentCommit.Hash)
			// set all files in tree to current commit time
			fileIter := currentTree.Files()
			for {
				file, err := fileIter.Next()
				if err != nil {
					slog.Debug("file iter done", "err", err)
					break
				}
				m.handleDiffPath(file.Name, currentCommitTime)
			}
			break
		}
		parentTree, err := parentCommit.Tree()
		if err != nil {
			return fmt.Errorf("get parent tree: %w", err)
		}

		changes, err := object.DiffTree(parentTree, currentTree)
		if err != nil {
			return fmt.Errorf("diff tree: %w", err)
		}

		for _, change := range changes {
			path := change.To.Name
			m.handleDiffPath(path, currentCommitTime)
		}
		currentCommit = parentCommit
		if currentCommitTime.Before(m.minCommitTime) {
			m.minCommitTime = currentCommitTime
		}
		depth += 1
		if depth >= m.maxDepth {
			slog.Info("reached max depth", "depth", depth)
			return nil
		}
	}
	return nil
}

func (m *Manager) UpdateFilesystem() error {
	updateCtr := 0
	err := filepath.Walk(m.repoRoot, func(path string, info os.FileInfo, _ error) error {
		relPath, err := filepath.Rel(m.repoRoot, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		if strings.HasPrefix(relPath, ".git") {
			return nil
		}
		if m.ignore != nil && m.ignore.MatchesPath(relPath) {
			return nil
		}
		if relPath == "." {
			return nil
		}
		lastCommitTime, ok := m.modTimes[relPath]
		if !ok {
			if m.maxDepth == math.MaxInt64 {
				slog.Warn("path not found in commit history", "path", relPath)
				return nil
			}
			// use minCommitTime if not found
			lastCommitTime = m.minCommitTime
		}
		if info.ModTime().Equal(lastCommitTime) {
			slog.Debug("modtime equal commit time", "path", relPath)
			return nil
		}
		slog.Debug("setting mod time", "path", relPath, "oldTime", info.ModTime(), "newTime", lastCommitTime)
		err = os.Chtimes(path, lastCommitTime, lastCommitTime)
		if err != nil {
			return err
		}
		updateCtr += 1
		return nil
	})
	slog.Info("updated mod times", "count", updateCtr)
	return err
}

func (m *Manager) RunDefault() error {
	err := m.SetFromGit()
	if err != nil {
		return fmt.Errorf("populate max commit times: %w", err)
	}
	err = m.UpdateFilesystem()
	if err != nil {
		return fmt.Errorf("update filesystem: %w", err)
	}
	return nil
}
