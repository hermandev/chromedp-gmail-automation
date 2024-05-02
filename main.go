package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

type User struct {
	Email    string
	Password string
}

func main() {
	runtime.GOMAXPROCS(2)
	// Open file csv
	f, err := os.Open("data.csv")
	if err != nil {
		log.Fatal(err)
		return
	}

	defer f.Close()

	csvRead := csv.NewReader(f)

	reader, err := csvRead.ReadAll()
	if err != nil {
		log.Fatal(err)
		return
	}

	var users []User

	for _, data := range reader {
		user := User{
			Email:    data[0],
			Password: data[1],
		}
		users = append(users, user)
	}

	OpenBrowser(users)
}

func OpenBrowser(data []User) {
	var wg sync.WaitGroup
	done := make(chan bool)

	for _, user := range data {
		wg.Add(1)

		go func(user User) {
			fmt.Printf("Email: %s", user.Email)
			opts := append(chromedp.DefaultExecAllocatorOptions[:],
				chromedp.DisableGPU,
				chromedp.UserDataDir("tmp/"+user.Email),
				chromedp.Flag("headless", false),
				chromedp.Flag("enable-automation", false),
				chromedp.Flag("incognito", true),
				chromedp.Flag("restore-on-startup", false),
			)

			allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

			ctx, _ := chromedp.NewContext(allocCtx)

			if err := chromedp.Run(ctx, chromedp.Tasks{
				runTask(chromedp.Navigate("https://accounts.google.com/"), "Open Google Account"),
				runTask(chromedp.WaitVisible("#identifierId", chromedp.ByID), "Check Input Email"),
				runTask(chromedp.SendKeys("#identifierId", user.Email, chromedp.ByID), "Input Email"),
				runTask(chromedp.Click("#identifierNext", chromedp.ByID), "Next"),
				runTask(chromedp.WaitVisible("//*[@id='password']/div[1]/div/div[1]/input", chromedp.BySearch), "Check Input Password"),
				runTask(chromedp.SendKeys("//*[@id='password']/div[1]/div/div[1]/input", user.Password, chromedp.BySearch), "Input Password"),
				runTask(chromedp.Click("/html/body/div[1]/div[1]/div[2]/c-wiz/div/div[3]/div/div[1]/div/div/button", chromedp.BySearch), "Next Login"),
				// aktifkan comment ini jika meminta confirm
				// runTask(chromedp.WaitVisible("//*[@id='tos_form']", chromedp.BySearch), "Wait Confirm "),
				// runTask(chromedp.Submit("//*[@id='tos_form']", chromedp.BySearch), "Submit Confirm"),
				// chromedp.WaitNotVisible("//*[@id='tos_form']", chromedp.BySearch),
				runTask(chromedp.WaitVisible("/html/body/div[4]/header", chromedp.BySearch), "Wait Profile"),
				runTask(chromedp.Navigate("http://tiny.cc/jokoganteng"), "Open Cloud Console"),
				runTask(chromedp.WaitVisible("/html/body/div/div[2]/div", chromedp.BySearch), "Wait Modal"),
				runTask(chromedp.Click("//*[@id='mat-mdc-checkbox-1-input']", chromedp.BySearch), "Checklis Modal"),
				runTask(chromedp.Click("/html/body/div/div[2]/div/mat-dialog-container/dialog-overlay/div[5]/modal-action/button", chromedp.BySearch), "Config Modal"),
				runTask(chromedp.WaitNotVisible("//*[@id='cloudshell']/div/loading-screen/div/div/div[1]/div", chromedp.BySearch), "Wait Close Modal"),
				runTask(chromedp.Click("document.querySelector('#cloudshell > standalone-header > div > mat-toolbar > span > cloudshell-view-controls > visibility-toggle:nth-child(2)')", chromedp.ByJSPath), "Click Button Terminal"),
			}); err != nil {
				log.Fatalln(err)
			}

			time.Sleep(time.Minute)
			close(done)
		}(user)
	}

	wg.Wait()
	<-done
}

func runTask(cdp chromedp.QueryAction, str string) chromedp.QueryAction {
	fmt.Println(str)
	return cdp
}
