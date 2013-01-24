package crawler

import (
    "io/ioutil"
    "net/http"
    "net/url"
)

type Jar struct {
	cookies []*http.Cookie
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	client := http.Client{nil, nil, new(Jar)}
	return &HttpClient{&client}
}

func (h *HttpClient) Get(url string) ([]byte, error) {
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (h *HttpClient) Post(url string, data map[string][]string) ([]byte, error) {
	resp, err := h.client.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if location, _ := resp.Location(); location != nil {
		return h.Get(location.String())
	}

	return ioutil.ReadAll(resp.Body)
}
