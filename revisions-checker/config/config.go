package config

type Configuration struct {
	ProjectID                                 string
	ServiceName                               string
	Region                                    string
	SlackWebhookURL                           string
	MostRecentRevisionsFirebaseDocumentPrefix string
	ActiveRevisionsFirebaseDocumentPrefix     string
}
