package config

func Init(cfg *Config) error {
	for _, repo := range cfg.Repos {
		for _, event := range repo.Events {
			for _, match := range event.Matches {
				if err := match.Compile(); err != nil {
					return err
				}
				for _, ev := range match.Events {
					if ev.Name == "pull_request" && ev.Types == nil {
						// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
						// > By default, a workflow only runs when a pull_request event's activity type is
						// > opened, synchronize, or reopened.
						ev.Types = []string{"opened", "synchronize", "reopened"}
					}
				}
			}
		}
	}
	return nil
}
