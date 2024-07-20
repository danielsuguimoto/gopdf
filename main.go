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
	app.Listen(":3000")
}

func handleFromUrl(c *fiber.Ctx) error {
	url := c.Query("url")

	if len(url) == 0 {
		return c.Status(400).JSON(map[string]string{"msg": "missing parameter: url"})
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, printUrlToPDF(url, &buf)); err != nil {
		log.Fatal(err)
	}

	return deliverPdfFile(buf, c)
}

func handleFromHtml(c *fiber.Ctx) error {
	r := new(FromHtmlRequest)
	if err := c.BodyParser(r); err != nil {
		return err
	}

	html := r.Html

	filepath := uuid.NewString() + ".html"

	if err := saveHtmlFile(filepath, html); err != nil {
		return err
	}

	defer os.Remove(filepath)

	return c.Status(200).SendFile(filepath, true)
}

func deliverPdfFile(res []byte, c *fiber.Ctx) error {
	fileName := uuid.New().String() + ".pdf"
	newFile, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer newFile.Close()
	newFile.Write(res)

	defer os.Remove(fileName)

	return c.Status(200).SendFile(fileName, true)
}

func saveHtmlFile(filepath string, html string) error {
	newFile, err := os.Create(filepath)

	if err != nil {
		return err
	}

	defer newFile.Close()
	newFile.WriteString(html)
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