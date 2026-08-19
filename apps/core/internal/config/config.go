package config

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

type DatabaseConfig struct { Driver string; Host string; Port int; Database string; Username string; Password string; AutoMigrate bool }
type LoggerConfig struct { Level string; Format string }
type Provider interface { Load(context.Context, []string) (map[string]string, error) }
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Logger      LoggerConfig
	JWT         JWTConfig
	OAuth       OAuthConfig
	SMTP        SMTPConfig
	R2          R2Config
	Gemini      GeminiConfig
	FrontendURL string
}

type GeminiConfig struct { Backend string; APIKey string; Model string; TimeoutSeconds int }
type SMTPConfig struct { Host string; Port int; Username string; Password string; From string; UseTLS bool }
type R2Config struct { AccountID string; AccessKeyID string; SecretAccessKey string; BucketName string; PublicURL string }
type OAuthConfig struct { Google GoogleOAuthConfig }
type GoogleOAuthConfig struct { ClientID string; ClientSecret string }
type JWTConfig struct { Secret string; ExpireHour int; RefreshExpireHour int }
type ServerConfig struct { Address string; Port int; Mode string; APIBasePath string }
type cloudflareKVProvider struct { client *http.Client; accountID, namespaceID, token, prefix, baseURL string }

const cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

func newCloudflareKVProvider() Provider {
	account, namespace, token := os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_KV_NAMESPACE_ID"), os.Getenv("CLOUDFLARE_API_TOKEN")
	if account == "" || namespace == "" || token == "" { return nil }
	return &cloudflareKVProvider{http.DefaultClient, account, namespace, token, os.Getenv("CLOUDFLARE_KV_PREFIX"), cloudflareAPIBase}
}
func (p *cloudflareKVProvider) Load(ctx context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string)
	apiBase := p.baseURL
	if apiBase == "" { apiBase = cloudflareAPIBase }
	base := apiBase + "/accounts/" + p.accountID + "/storage/kv/namespaces/" + p.namespaceID + "/values/"

	// Prefixed keys first (e.g. cluster/DB_HOST). Bulk-get them in one request.
	prefixed := make([]string, 0, len(keys))
	for _, k := range keys {
		if p.prefix != "" { prefixed = append(prefixed, p.prefix+k) }
	}
	if len(prefixed) > 0 {
		got, err := p.loadKeysBulk(ctx, base, prefixed)
		if err != nil { return nil, err }
		for k, v := range got {
			values[strings.TrimPrefix(k, p.prefix)] = v
		}
	}

	// Fall back to unprefixed keys for anything not found above.
	var missing []string
	for _, k := range keys {
		if _, ok := values[k]; !ok { missing = append(missing, k) }
	}
	if len(missing) > 0 {
		got, err := p.loadKeysBulk(ctx, base, missing)
		if err != nil { return nil, err }
		for k, v := range got { values[k] = v }
	}
	return values, nil
}

// loadKeysBulk fetches multiple keys in one Cloudflare KV bulk-get call,
// returning only the keys that exist.
func (p *cloudflareKVProvider) loadKeysBulk(ctx context.Context, base string, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	body, err := json.Marshal(map[string][]string{"keys": keys})
	if err != nil { return nil, err }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"bulk/get", bytes.NewReader(body))
	if err != nil { return nil, err }
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudflare KV bulk/get returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Result struct {
			Values map[string]string `json:"values"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil { return nil, err }
	for k, v := range parsed.Result.Values { result[k] = v }
	return result, nil
}

func Load() (*Config, error) {
	keys := []string{"SERVER_ADDRESS","SERVER_PORT","SERVER_MODE","API_BASE_PATH","DB_DRIVER","DB_HOST","DB_PORT","DB_NAME","DB_USER","DB_PASSWORD","DB_AUTO_MIGRATE","LOG_LEVEL","LOG_FORMAT","JWT_SECRET","JWT_EXPIRE_HOUR","JWT_REFRESH_EXPIRE_HOUR","GOOGLE_CLIENT_ID","GOOGLE_CLIENT_SECRET","FRONTEND_URL","SMTP_HOST","SMTP_PORT","SMTP_USERNAME","SMTP_PASSWORD","SMTP_FROM","SMTP_USE_TLS","R2_ACCOUNT_ID","R2_ACCESS_KEY_ID","R2_SECRET_ACCESS_KEY","R2_BUCKET_NAME","R2_PUBLIC_URL","GEMINI_BACKEND","GEMINI_API_KEY","GEMINI_MODEL","GEMINI_TIMEOUT_SECONDS"}
	values := make(map[string]string, len(keys))
	for _, key := range keys { if value, ok := os.LookupEnv(key); ok { values[key] = value } }
	if provider := newCloudflareKVProvider(); provider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), providerTimeout()); defer cancel()
		if remote, err := provider.Load(ctx, keys); err == nil {
			for key, value := range remote { values[key] = value }
		}
	}
	return decode(values)
}

func providerTimeout() time.Duration { if d, err := time.ParseDuration(os.Getenv("CLOUDFLARE_KV_TIMEOUT")); err == nil && d > 0 { return d }; return 5*time.Second }
func get(v map[string]string, key, fallback string) string { if value, ok := v[key]; ok { return value }; return fallback }
func integer(v map[string]string, key string, fallback int) (int, error) { raw:=get(v,key,strconv.Itoa(fallback)); n,err:=strconv.Atoi(raw); if err!=nil{return 0,fmt.Errorf("%s: %w",key,err)};return n,nil }
func boolean(v map[string]string, key string, fallback bool) (bool,error) { raw:=get(v,key,strconv.FormatBool(fallback)); b,err:=strconv.ParseBool(raw);if err!=nil{return false,fmt.Errorf("%s: %w",key,err)};return b,nil }
func decode(v map[string]string) (*Config,error) {
	var err error; c:=&Config{}
	c.Server.Address=get(v,"SERVER_ADDRESS",":8080"); c.Server.Port,err=integer(v,"SERVER_PORT",8080);if err!=nil{return nil,err};c.Server.Mode=get(v,"SERVER_MODE","debug");c.Server.APIBasePath=get(v,"API_BASE_PATH","/api/v1")
	c.Database.Driver=get(v,"DB_DRIVER","postgres");c.Database.Host=get(v,"DB_HOST","localhost");c.Database.Port,err=integer(v,"DB_PORT",5432);if err!=nil{return nil,err};c.Database.Database=get(v,"DB_NAME","revieu");c.Database.Username=get(v,"DB_USER","postgres");c.Database.Password=get(v,"DB_PASSWORD","");c.Database.AutoMigrate,err=boolean(v,"DB_AUTO_MIGRATE",false);if err!=nil{return nil,err}
	c.Logger.Level=get(v,"LOG_LEVEL","info");c.Logger.Format=get(v,"LOG_FORMAT","json");c.JWT.Secret=get(v,"JWT_SECRET","");c.JWT.ExpireHour,err=integer(v,"JWT_EXPIRE_HOUR",24);if err!=nil{return nil,err};c.JWT.RefreshExpireHour,err=integer(v,"JWT_REFRESH_EXPIRE_HOUR",168);if err!=nil{return nil,err}
	c.OAuth.Google.ClientID=get(v,"GOOGLE_CLIENT_ID","");c.OAuth.Google.ClientSecret=get(v,"GOOGLE_CLIENT_SECRET","");c.FrontendURL=get(v,"FRONTEND_URL","")
	c.SMTP.Host=get(v,"SMTP_HOST","smtp.gmail.com");c.SMTP.Port,err=integer(v,"SMTP_PORT",587);if err!=nil{return nil,err};c.SMTP.Username=get(v,"SMTP_USERNAME","");c.SMTP.Password=get(v,"SMTP_PASSWORD","");c.SMTP.From=get(v,"SMTP_FROM","");c.SMTP.UseTLS,err=boolean(v,"SMTP_USE_TLS",true);if err!=nil{return nil,err}
	c.R2.AccountID=get(v,"R2_ACCOUNT_ID","");c.R2.AccessKeyID=get(v,"R2_ACCESS_KEY_ID","");c.R2.SecretAccessKey=get(v,"R2_SECRET_ACCESS_KEY","");c.R2.BucketName=get(v,"R2_BUCKET_NAME","revieu-media");c.R2.PublicURL=get(v,"R2_PUBLIC_URL","")
	c.Gemini.Backend=get(v,"GEMINI_BACKEND","gemini-api");c.Gemini.APIKey=get(v,"GEMINI_API_KEY","");c.Gemini.Model=get(v,"GEMINI_MODEL","gemini-2.5-flash");c.Gemini.TimeoutSeconds,err=integer(v,"GEMINI_TIMEOUT_SECONDS",30);if err!=nil{return nil,err}
	if strings.TrimSpace(c.JWT.Secret)=="" { return nil,fmt.Errorf("JWT_SECRET is required") }; return c,nil
}
