	package main

	import (
		"errors"
		"flag"
		"io"
		"log"
		"os"

		"github.com/go-git/go-billy/v5/memfs"
		"github.com/go-git/go-git/v5"
		"github.com/go-git/go-git/v5/plumbing"
		"github.com/go-git/go-git/v5/plumbing/transport/ssh"
		"github.com/go-git/go-git/v5/storage/memory"
		stdssh "golang.org/x/crypto/ssh"
	)

	const (
		repo_addr_const = "https://github.com/RYunisov/atomic-notes.git"
		refs_const      = "refs/heads/main"
		file_path_const = "README.md"
	)

	var username, private_key, refs, commit_hash, repo_addr, file_path string
	var skip_auth bool

	func init() {
		usr := os.Getenv("USER")
		pk := "/home/"  + usr + "/.ssh/id_rsa"

		flag.StringVar(&username, "user", usr, "Example Username")
		flag.StringVar(&private_key, "pk", pk, "Example PrivateKey")
		flag.StringVar(&refs, "refs", refs_const, "Example Refs")
		flag.StringVar(&commit_hash, "commit", "", "Example CommitId")
		flag.StringVar(&repo_addr, "repo", repo_addr_const, "Example RepoAddr")
		flag.StringVar(&file_path, "file", file_path_const, "Example FilePath")
		flag.BoolVar(&skip_auth, "skip-auth", true, "Skip Auth by Default")
	}

	func main() {

		fs := memfs.New()

		if err := checkFlags(os.Args[1:]); err != nil {
			// log.Fatalf("ERROR: checkFlags: %s", err)
			log.Print("Using args by default")
		}

		flag.Parse()

		co := configure()

		repo, err := git.Clone(memory.NewStorage(), fs, co)
		if err != nil {
			log.Fatalf("ERROR: Clone: %s", err)
		}

		if commit_hash != "" {
			w, err := repo.Worktree()
			if err != nil {
				log.Fatalf("ERROR: Worktree: %s", err)
			}

			err = w.Checkout(&git.CheckoutOptions{
					Hash: plumbing.NewHash(commit_hash),
			})
			if err != nil {
				log.Fatalf("ERROR: Checkout: %s", err)
			}
		}

		file, err := fs.Open(file_path)
		if err != nil {
			log.Fatalf("ERROR: Open: %s %s", file_path, err)
		}

		io.Copy(os.Stdout, file)
	}

	func checkFlags(args []string) error {
		if len(args) < 1 {
			return errors.New("arguments didn't passed")
		}
		return nil
	}

	func configure() *git.CloneOptions {
		if skip_auth {
			if commit_hash != "" {
				return &git.CloneOptions{
					URL:           repo_addr,
					Auth:          nil, 
					ReferenceName: plumbing.ReferenceName(refs),
					SingleBranch:  true,
				}	
			}
			return &git.CloneOptions{
				URL:           repo_addr,
				Auth:          nil, 
				ReferenceName: plumbing.ReferenceName(refs),
				SingleBranch:  true,
				Depth:         1,
			}
		}
		sshKey, _ := os.ReadFile(private_key)
		s, err := stdssh.ParsePrivateKey(sshKey)
		if err != nil {
			log.Print(err)
		}

		a := &ssh.PublicKeys{
			User:   username,
			Signer: s,
		}
		a.HostKeyCallback = stdssh.InsecureIgnoreHostKey()

		if commit_hash != "" {
			return &git.CloneOptions{
				URL:           repo_addr,
				Auth:          a, 
				ReferenceName: plumbing.ReferenceName(refs),
				SingleBranch:  true,
			}	
		}

		return &git.CloneOptions{
			URL:           repo_addr,
			Auth:          a,
			ReferenceName: plumbing.ReferenceName(refs),
			SingleBranch:  true,
			Depth:         1,
		}

	}
