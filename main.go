package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
)

type Cfg struct {
	WorkingDir        string
	TelegramToken     string
	CronSpec          string
	WhitelistedChatID string
	CronJitterSec     int64
}

var cfg Cfg

func initCfg() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	log.Printf("Seed: %d\n", seed)

	var err error
	cfg.WorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.WorkingDir[len(cfg.WorkingDir)-1] != '/' {
		cfg.WorkingDir += "/"
	}

	log.Println("Working directory:", cfg.WorkingDir)

	cfg.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}
	os.Setenv("TELEGRAM_TOKEN", "")

	cfg.WhitelistedChatID = os.Getenv("TELEGRAM_CHATID")
	if cfg.WhitelistedChatID == "" {
		log.Fatal("TELEGRAM_CHATID is not set")
	}
	log.Printf("Whitelisted chat ID: %s", cfg.WhitelistedChatID)

	cfg.CronSpec = os.Getenv("CRON_SPEC")
	cfg.CronJitterSec = getEnvInt("CRON_JITTER_SEC", 0)
}

const MAX_DEPTH = 32

func storeIndex(dir string) {
	log.Println("storeIndex: >>")
	defer log.Println("storeIndex: <<")

	var index []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				index = append(index, path)
			}
		}

		return nil
	})

	if err != nil {
		log.Println("storeIndex: Walk Error:", err)
		return
	}

	indexTmp := filepath.Join(dir, "index.txt.tmp")
	err = ioutil.WriteFile(indexTmp, []byte(strings.Join(index, "\n")), 0644)
	if err != nil {
		log.Println("storeIndex: Write Error:", err)
		return
	}

	indexFile := filepath.Join(dir, "index.txt")
	err = os.Rename(indexTmp, indexFile)
	if err != nil {
		log.Println("storeIndex: Move Error:", err)
		return
	}
}

func loadIndex(dir string) ([]string, error) {
	log.Println("loadIndex: >>")
	defer log.Println("loadIndex: <<")

	indexFile := filepath.Join(dir, "index.txt")
	data, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func randChooseFile(dir string) (string, error) {
	log.Println("RandChooseFile: >>")
	defer log.Println("RandChooseFile: <<")

	files, err := loadIndex(dir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files in %s", dir)
	}

	return files[rand.Intn(len(files))], nil
}

func getEnvInt(name string, defaultValue int64) int64 {
	vstr := os.Getenv(name)
	if vstr != "" {
		vint, err := strconv.ParseInt(vstr, 10, 64)
		if err != nil {
			log.Fatalf("%s is not a valid number: err=%s",
				name, err.Error())
		}
		return vint
	}
	return defaultValue
}

func handleHelp(c tele.Context) error {
	log.Println("handleHelp: >>")
	defer log.Println("handleHelp: <<")
	msg := "The bot for Lunohod community.\n"
	msg += fmt.Sprintf("Chat ID: %d\nCommands:\n", c.Chat().ID)
	msg += "/help - show this help\n"
	msg += "/photo - random photo of community\n"
	msg += "/beer - implement me\n"
	return c.Reply(msg)
}

func createRandomPhoto() (*tele.Photo, error) {
	file, err := randChooseFile(cfg.WorkingDir)
	if err != nil {
		return nil, err
	}
	log.Println("File:", file)

	caption := file[len(cfg.WorkingDir):]
	return &tele.Photo{File: tele.FromDisk(file), Caption: caption}, nil
}

func getRecipient(c tele.Context) string {
	r := c.Recipient()
	var s = ""
	if r != nil {
		s = r.Recipient()
	}
	return s
}

func VerifyAccess(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		recipient := getRecipient(c)
		if recipient != cfg.WhitelistedChatID {
			log.Printf("Not whitelisted '%s'. Access denied. \n", recipient)
			return c.Reply("Access denied")
		}
		return next(c)
	}
}

func handlePhoto(c tele.Context) error {
	log.Println("handlePhoto: >>")
	defer log.Println("handlePhoto: <<")

	photo, err := createRandomPhoto()
	if err != nil {
		log.Println("handlePhoto: Error:", err)
		return err
	}
	return c.Reply(photo)
}

func handleBeer(c tele.Context) error {
	log.Println("handleBeer: >>")
	defer log.Println("handleBeer: <<")

	return c.Reply("still no beer 🍺")
}

type Recipient struct {
	ChatID string
}

func (r *Recipient) Recipient() string {
	return r.ChatID
}

func handleCron(b *tele.Bot) {
	log.Println("handleCron: >>")
	defer log.Println("handleCron: <<")

	if cfg.CronJitterSec > 0 {
		jitter := rand.Int63n(cfg.CronJitterSec)
		log.Printf("Jitter: %d\n", jitter)
		time.Sleep(time.Duration(jitter) * time.Second)
	}

	photo, err := createRandomPhoto()
	if err != nil {
		log.Println("handleCron: Error:", err)
		return
	}
	photo.Caption = "Photo of the day: " + photo.Caption
	photo.Send(b, &Recipient{ChatID: cfg.WhitelistedChatID}, &tele.SendOptions{})
}

func main() {
	initCfg()
	storeIndex(cfg.WorkingDir)

	pref := tele.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Use(VerifyAccess)

	b.Handle("/help", handleHelp)
	b.Handle("/photo", handlePhoto)
	b.Handle("/beer", handleBeer)

	if cfg.CronSpec != "" {
		c := cron.New(
			cron.WithLogger(
				cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))

		c.AddFunc(cfg.CronSpec, func() {
			handleCron(b)
		})

		c.AddFunc("0 * * * *", func() {
			storeIndex(cfg.WorkingDir)
		})

		log.Printf("Cron spec: %s", cfg.CronSpec)
		c.Start()
		defer c.Stop()
	}

	log.Println("Bot started")
	b.Start()
}
