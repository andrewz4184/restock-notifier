# Example: How to add LLM to your existing system

## Modified HandleRequest function:

```go
func HandleRequest(ctx context.Context) (string, error) {
    // ... existing scraping logic ...
    
    if len(inStockItems) > 0 {
        if state == nil {
            // Instead of just notifying, consult LLM
            decision, err := consultLLMForPurchase(checker.products)
            if err != nil {
                // Fallback to notification
                return sendNotificationOnly(checker.FormatMessage())
            }
            
            if decision.ShouldPurchase {
                // Execute purchase
                result, err := executePurchaseWorkflow(decision)
                if err != nil {
                    return fmt.Sprintf("Purchase failed: %v", err)
                }
                return fmt.Sprintf("Successfully purchased: %s", result.OrderID)
            } else {
                // LLM decided not to purchase
                return sendNotificationOnly(decision.Reason)
            }
        }
    }
    
    // ... rest of existing logic ...
}

type PurchaseDecision struct {
    ShouldPurchase   bool
    SelectedProduct  Product
    Reason          string
    MaxPrice        int
}

func consultLLMForPurchase(products []Product) (*PurchaseDecision, error) {
    prompt := buildPurchasePrompt(products)
    
    // Call Claude API
    response, err := callClaudeAPI(prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse Claude's decision
    return parsePurchaseDecision(response)
}
```

## The LLM Prompt:

```
You are a matcha purchasing assistant. Analyze these available products:

${productList}

User preferences:
- Budget: $100 max per order
- Prefers: Wako or Aoarashi grades
- Avoid: Mukashi (too strong)

Current inventory: Has 2 cans of Wako already

Should I purchase? If yes, which product and why?
Respond in JSON: {"shouldPurchase": bool, "product": "name", "reason": "explanation"}
```

## Browser Automation Integration:

```go
func executePurchaseWorkflow(decision *PurchaseDecision) (*PurchaseResult, error) {
    // Start browser automation
    browser := playwright.NewBrowser()
    page := browser.NewPage()
    
    // Navigate and add to cart
    err := addToCart(page, decision.SelectedProduct)
    if err != nil {
        // If error, ask LLM for help
        solution := askLLMToSolveCheckoutIssue(err, page.Screenshot())
        err = applySolution(page, solution)
    }
    
    // Continue checkout with LLM guidance
    return completeCheckout(page, decision)
}
```
