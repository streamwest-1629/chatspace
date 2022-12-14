package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/streamwest-1629/chatspace/app/talker"
	"github.com/streamwest-1629/chatspace/app/voicevox"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewDevelopment()
}

func main() {

	// healthcheck
	logger.Debug("initialize application")
	http.HandleFunc("/_hck", healthcheck)

	config := voicevox.InitConfig{
		NumThreads:    2,
		LoadAllModels: true,
	}

	// voicevox application
	vv, err := voicevox.Start(logger, os.Getenv("VOICEVOX_COREPATH"), os.Getenv("VOICEVOX_JTALKDIR"), config)
	if err != nil {
		logger.Fatal("cannot start voicevox application", zap.Error(err))
	}
	defer vv.Quit()
	time.Sleep(200 * time.Millisecond)
	if _, err := vv.GetSpeakers("", true); err != nil {
		logger.Fatal("cannot start voicevox application", zap.Error(err))
	}

	// chatspace application
	discordToken := os.Getenv("TALKER_DISCORD_TOKEN")
	controller, err := talker.NewService(logger, discordToken, vv)
	if err != nil {
		logger.Error("cannot start talker application", zap.String("discordToken", discordToken[:8]+"***"+discordToken[len(discordToken)-8:]), zap.Error(err))
	}
	defer controller.Close()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("listen server is ended", zap.Error(err))
	}
}

func healthcheck(rw http.ResponseWriter, req *http.Request) {
	logger.Debug("call healthcheck")

	current := time.Now().UTC().Format(time.RFC3339)

	b, _ := json.Marshal(map[string]string{
		"currentTime": current,
	})

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b)
}
