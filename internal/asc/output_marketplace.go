package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func printMarketplaceSearchDetailsTable(resp *MarketplaceSearchDetailsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCatalog URL")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.CatalogURL),
		)
	}
	return w.Flush()
}

func printMarketplaceSearchDetailsMarkdown(resp *MarketplaceSearchDetailsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Catalog URL |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.CatalogURL),
		)
	}
	return nil
}

func printMarketplaceSearchDetailTable(resp *MarketplaceSearchDetailResponse) error {
	return printMarketplaceSearchDetailsTable(&MarketplaceSearchDetailsResponse{
		Data: []Resource[MarketplaceSearchDetailAttributes]{resp.Data},
	})
}

func printMarketplaceSearchDetailMarkdown(resp *MarketplaceSearchDetailResponse) error {
	return printMarketplaceSearchDetailsMarkdown(&MarketplaceSearchDetailsResponse{
		Data: []Resource[MarketplaceSearchDetailAttributes]{resp.Data},
	})
}

func printMarketplaceWebhooksTable(resp *MarketplaceWebhooksResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEndpoint URL")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.EndpointURL),
		)
	}
	return w.Flush()
}

func printMarketplaceWebhooksMarkdown(resp *MarketplaceWebhooksResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Endpoint URL |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.EndpointURL),
		)
	}
	return nil
}

func printMarketplaceWebhookTable(resp *MarketplaceWebhookResponse) error {
	return printMarketplaceWebhooksTable(&MarketplaceWebhooksResponse{
		Data: []Resource[MarketplaceWebhookAttributes]{resp.Data},
	})
}

func printMarketplaceWebhookMarkdown(resp *MarketplaceWebhookResponse) error {
	return printMarketplaceWebhooksMarkdown(&MarketplaceWebhooksResponse{
		Data: []Resource[MarketplaceWebhookAttributes]{resp.Data},
	})
}

func printMarketplaceSearchDetailDeleteResultTable(result *MarketplaceSearchDetailDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printMarketplaceSearchDetailDeleteResultMarkdown(result *MarketplaceSearchDetailDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printMarketplaceWebhookDeleteResultTable(result *MarketplaceWebhookDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printMarketplaceWebhookDeleteResultMarkdown(result *MarketplaceWebhookDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}
