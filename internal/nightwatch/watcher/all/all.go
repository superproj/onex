package all

//nolint: golint
import (
	_ "github.com/superproj/onex/internal/nightwatch/watcher/cronjob/cronjob"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/cronjob/statesync"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/job/llmtrain"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/secretsclean"
	_ "github.com/superproj/onex/internal/nightwatch/watcher/user"
)
