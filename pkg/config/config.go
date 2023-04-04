package config

import "time"

type Config struct {
	Store      Store
	SSH        SSH
	ConfigRepo ConfigRepo
}

type Store struct {
	DatabaseFile string
	GitDir       string
}

type SSH struct {
	KeyDir string
}

type ConfigRepo struct {
	CloneURL                string
	BranchName              string
	ReconciliationIntervall time.Duration
}

var Cfg Config

func init() {
	// set defaults
	Cfg = Config{
		Store: Store{
			DatabaseFile: "./.data/bbolt.db",
			GitDir:       "./.data/repos/",
		},
		SSH: SSH{
			KeyDir: "./.data/",
		},
		ConfigRepo: ConfigRepo{
			CloneURL:                "git@github.com:LaCodon/recoon-test.git",
			BranchName:              "test",
			ReconciliationIntervall: 5 * time.Second,
		},
	}
}
