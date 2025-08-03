package main

// Updated by Claude for live notification management
import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gocolly/colly/v2"
)

type Product struct {
	Title  string
	Status string
}

type NotificationState struct {
	Date         string    `json:"date"`
	NotifiedAt   time.Time `json:"notified_at"`
	InStockItems []string  `json:"in_stock_items"`
}

type StockChecker struct {
	products  []Product
	collector *colly.Collector
	dynamoDB  *dynamodb.Client
	tableName string
}

func NewStockChecker() *StockChecker {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	ddb := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("DYNAMODB_TABLE")

	return &StockChecker{
		collector: colly.NewCollector(
			colly.AllowedDomains("www.marukyu-koyamaen.co.jp"),
		),
		dynamoDB:  ddb,
		tableName: tableName,
	}
}

func (sc *StockChecker) ScrapeProducts() error {
	sc.collector.OnHTML("li.product", func(e *colly.HTMLElement) {
		title := e.ChildAttr("a.woocommerce-loop-product__link", "title")
		status := "❌ Out of Stock"
		if !strings.Contains(e.Attr("class"), "outofstock") {
			status = "✅ In Stock"
		}
		sc.products = append(sc.products, Product{Title: title, Status: status})
	})

	return sc.collector.Visit("https://www.marukyu-koyamaen.co.jp/english/shop/products/catalog/matcha/principal")
}

func (sc *StockChecker) GetNotificationState(date string) (*NotificationState, error) {
	result, err := sc.dynamoDB.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &sc.tableName,
		Key: map[string]types.AttributeValue{
			"date": &types.AttributeValueMemberS{Value: date},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		// No notification sent today
		return nil, nil
	}

	var state NotificationState
	err = attributevalue.UnmarshalMap(result.Item, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (sc *StockChecker) UpdateNotificationState(date string, inStockItems []string) error {
	state := NotificationState{
		Date:         date,
		NotifiedAt:   time.Now(),
		InStockItems: inStockItems,
	}

	item, err := attributevalue.MarshalMap(state)
	if err != nil {
		return err
	}

	_, err = sc.dynamoDB.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &sc.tableName,
		Item:      item,
	})

	return err
}

func (sc *StockChecker) ClearNotificationState(date string) error {
	_, err := sc.dynamoDB.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &sc.tableName,
		Key: map[string]types.AttributeValue{
			"date": &types.AttributeValueMemberS{Value: date},
		},
	})
	return err
}

func (sc *StockChecker) FormatMessage() string {
	japanLocation, _ := time.LoadLocation("Asia/Tokyo")
	japanTime := time.Now().In(japanLocation)

	message := "Marukyu Koyamaen Stock Check:\n\n"
	message += fmt.Sprintf("🕜 Last Checked: %s (Japan Time)\n\n",
		japanTime.Format("Mon, 2 Jan 3:04 PM"))

	for _, product := range sc.products {
		message += fmt.Sprintf("🍵 Name: %s\n📦 Status: %s\n\n",
			product.Title,
			product.Status)
	}

	return message
}

func sendTelegramNotification(message string) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		return fmt.Errorf("telegram bot token or chat ID is not set")
	}

	telegramAPI := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	response, err := http.PostForm(
		telegramAPI,
		url.Values{
			"chat_id": {chatID},
			"text":    {message},
		})

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", response.Status)
	}

	return nil
}

func HandleRequest(ctx context.Context) (string, error) {
	fmt.Println("Starting matcha stock check...")
	checker := NewStockChecker()

	fmt.Println("Scraping products...")
	err := checker.ScrapeProducts()
	if err != nil {
		return "", fmt.Errorf("failed to scrape products: %v", err)
	}
	fmt.Printf("Found %d products\n", len(checker.products))

	// Get current date in JST for state tracking
	japanLocation, _ := time.LoadLocation("Asia/Tokyo")
	currentDate := time.Now().In(japanLocation).Format("2006-01-02")
	fmt.Printf("Current date (JST): %s\n", currentDate)

	// Check if any product is back in stock
	inStockItems := []string{}
	for _, product := range checker.products {
		if product.Status == "✅ In Stock" {
			inStockItems = append(inStockItems, product.Title)
		}
	}
	fmt.Printf("Products in stock: %d\n", len(inStockItems))

	// Get current notification state
	fmt.Println("Checking notification state...")
	state, err := checker.GetNotificationState(currentDate)
	if err != nil {
		return "", fmt.Errorf("failed to get notification state: %v", err)
	}

	if len(inStockItems) > 0 {
		// Products are in stock
		if state == nil {
			fmt.Println("Products in stock, sending first notification...")
			// Haven't notified today, send notification
			message := checker.FormatMessage()
			err = sendTelegramNotification(message)
			if err != nil {
				return "", fmt.Errorf("failed to send Telegram notification: %v", err)
			}

			// Update state to prevent duplicate notifications
			err = checker.UpdateNotificationState(currentDate, inStockItems)
			if err != nil {
				return "", fmt.Errorf("failed to update notification state: %v", err)
			}

			return "Notification sent", nil
		} else {
			fmt.Println("Products in stock, but already notified today")
			// Already notified today, do nothing
			return "Already notified today", nil
		}
	} else {
		// No products in stock
		if state != nil {
			fmt.Println("No products in stock, clearing notification state...")
			// Clear the notification state (ready for next restock)
			err = checker.ClearNotificationState(currentDate)
			if err != nil {
				return "", fmt.Errorf("failed to clear notification state: %v", err)
			}
			return "State cleared - ready for next restock", nil
		}
		fmt.Println("No products in stock, no action needed")
		return "No stock, no notification needed", nil
	}
}

func main() {
	lambda.Start(HandleRequest)
}
