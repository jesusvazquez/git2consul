package main

import (
	"errors"
	"github.com/thehivecorporation/log"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"strings"
)

var (
	ErrRepoAlreadyExists = errors.New("repository already exists")
	ErrRepoIsUpToDate    = errors.New("already up-to-date")
)

// Clones a repository to a given path in the filesystem
func cloneRepository(repository *string, directory *string, gitUser *string, gitPrivateKey *string) {
	// Private repo example in https://github.com/src-d/go-git/issues/377
	// Clone the given repository to the given directory
	options := &git.CloneOptions{}
	options.URL = *repository
	if len(*gitPrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey([]byte(strings.Replace(*gitPrivateKey, `\n`, "\n", -1)))
		if err != nil {
			log.WithError(err).Fatal("Can't parse given private key")
			panic(err)
		}
		auth := &ssh2.PublicKeys{User: *gitUser, Signer: signer}
		options.Auth = auth
	}
	_, err := git.PlainClone(*directory, false, options)
	if err != nil && err.Error() != ErrRepoAlreadyExists.Error() {
		log.WithError(err).Fatal("Error cloning repository " + *repository)
		panic(err)
	}

	log.Info("Repository already exists at HEAD: " + headCommit(directory))
}

// Pulls from origin the latest changes
func pullFromRepository(directory *string, gitUser *string, gitPrivateKey *string) *git.Worktree {
	// We instance a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(*directory)
	if err != nil {
		log.WithError(err).Fatal("Error opening local git directory " + *directory)
		panic(err)
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		log.WithError(err).Fatal("Error loading worktree")
		panic(err)
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	log.Debug("Pulling from origin")
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err.Error() != ErrRepoIsUpToDate.Error() {
		log.WithError(err).Fatal("Error pulling from repository")
		panic(err)
	}
	log.Info("Head Is at " + headCommit(directory))
	return w
}

// Returns the HEAD commit reference of a given git directory
func headCommit(directory *string) string {
	// Instance a new repository targeting the given path
	r, err := git.PlainOpen(*directory)
	if err != nil {
		log.WithError(err).Fatal("Error opening local git directory " + *directory)
		panic(err)
	}

	// Print the HEAD commit hash available in the directory
	ref, err := r.Head()
	if err != nil {
		log.WithError(err).Fatal("Error getting HEAD reference")
		panic(err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		log.WithError(err).Fatal("Error getting HEAD commit reference")
		panic(err)
	}

	return commit.Hash.String()
}
