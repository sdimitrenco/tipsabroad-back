package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/checkout/session"
	"log"
	"os"
	"strconv"
)

type createCheckoutSessionResponse struct {
	SessionID string `json:"id"`
}

type User struct {
	Id        int    `json:"user"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	CompanyId int    `json:"company_id"`
	Tip       int64  `json:"tip"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}

func CreateCheckoutSession(c *fiber.Ctx) error {
	var user User

	var images []*string

	bodyResponse := c.Body()

	err := json.Unmarshal(bodyResponse, &user)
	if err != nil {
		log.Printf("can't parse client responce, %v", err)
	}

	img := "https://gafki.ru/wp-content/uploads/2019/10/1-tipichnyj-dikij-lesnoj-kot.jpg" //demo

	if os.Getenv("PROD") == "yes" {
		img = fmt.Sprintf("%s/%s", os.Getenv("STATIC_IMG_HOST"), user.Image)
	}

	var description = &user.Comment

	if user.Comment == "" {
		description = nil
	}

	images = append(images, &img)

	successURL := fmt.Sprintf("?success=true&user=%d&tip=%d&company_id=%d", user.Id, user.Tip, user.CompanyId)
	canceledURL := fmt.Sprintf("?canceled=true&user=%d&tip=%d&company_id=%d", user.Id, user.Tip, user.CompanyId)

	//create metadata
	metaData := make(map[string]string)
	metaData["rating"] = strconv.Itoa(user.Rating)

	domain := os.Getenv("HOST")
	params := &stripe.CheckoutSessionParams{
		CancelURL: stripe.String(domain + canceledURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(user.Name),
						Images:      images,
						Metadata:    metaData,
						Description: description,
					},
					UnitAmount: stripe.Int64(user.Tip),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		SuccessURL: stripe.String(domain + successURL),
	}

	session, err := session.New(params)

	if err != nil {
		log.Printf("session.New: %v", err)
	}

	data := createCheckoutSessionResponse{
		SessionID: session.ID,
	}

	return c.JSON(data)

}
