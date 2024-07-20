package main

import (
	"context"
	"log"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FromHtmlRequest struct {
	Html string `json:"html"`
}

func main() {
	app := fiber.New()
	app.Get("/from-url", handleFromUrl)
	app.Post("/from-html", handleFromHtml)
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

func handleFromUrl(c *fiber.Ctx) (err error) {
	url := c.Query("url")

	if len(url) == 0 {
		return c.Status(400).JSON(map[string]string{"msg": "missing parameter: url"})
	}

	var buf []byte
	if buf, err = createPdfBufferFromUrl(url); err != nil {
		return err
	}

	return deliverPdfFile(buf, c)
}

func handleFromHtml(c *fiber.Ctx) (err error) {
	r := new(FromHtmlRequest)
	if err = c.BodyParser(r); err != nil {
		return err
	}

	html := r.Html

	filepath := "/tmp/" + uuid.NewString() + ".html"

	if err = saveHtmlFile(filepath, html); err != nil {
		return err
	}

	defer os.Remove(filepath)

	var buf []byte

	if buf, err = createPdfBufferFromUrl("file://" + filepath); err != nil {
		return err
	}

	return deliverPdfFile(buf, c)
}

func createPdfBufferFromUrl(url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, printUrlToPDF(url, &buf)); err != nil {
		return nil, err
	}

	return buf, nil
}

func deliverPdfFile(res []byte, c *fiber.Ctx) error {
	fileName := "/tmp/" + uuid.New().String() + ".pdf"
	newFile, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer newFile.Close()
	defer os.Remove(fileName)

	if _, err = newFile.Write(res); err != nil {
		return err
	}

	return c.Status(200).SendFile(fileName, true)
}

func saveHtmlFile(filepath string, html string) error {
	newFile, err := os.Create(filepath)

	if err != nil {
		return err
	}

	defer newFile.Close()
	if _, err = newFile.WriteString(html); err != nil {
		return err
	}
	return nil
}

func printUrlToPDF(url string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
