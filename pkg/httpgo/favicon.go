package httpgo

import (
	"fmt"
	"httpgo/pkg/utils"
	"net/url"
	"strings"
	"time"
)

type FaviconList struct {
	Url         string
	Favicon     []string
	FaviconHash []string
}

func (r *Response) GetFaviconHash(proxyStr string, timeoutInt time.Duration) (*FaviconList, error) {
	var favicons []string
	var faviconhash []string

	u, err := url.Parse(r.Url)
	if err != nil {
		return nil, err
	}

	baseURL := u.Scheme + "://" + u.Host
	mainFavicon := u.Scheme + "://" + u.Host + "/favicon.ico"
	favicons = append(favicons, mainFavicon)

	spareFavicon, err := utils.ExtractSpareFavicon(r.Body)
	if err != nil {
		return nil, err
	}
	for i := range spareFavicon {
		if strings.HasPrefix(spareFavicon[i], "http://") || strings.HasPrefix(spareFavicon[i], "https://") {
			favicons = append(favicons, spareFavicon[i])
		} else {
			fullURL, err := ResolveURL(baseURL, spareFavicon[i])
			if err != nil {
				favicons = append(favicons, mainFavicon)
				//log.Printf("Error resolving URL for %s: %v\n", baseURL, err)
			} else {
				favicons = append(favicons, fullURL)
			}
		}
	}
	favicons = RemoveDuplicates(favicons)

	for i := range favicons {
		fh, err := GetResponse(favicons[i], proxyStr, timeoutInt)
		if err != nil {
			return nil, err
		}
		faviconhash = append(faviconhash, utils.Mmh3Hash32([]byte(fh.Body)))
	}

	faviconhash = RemoveDuplicates(faviconhash)

	return &FaviconList{
		Url:         r.Url,
		Favicon:     favicons,
		FaviconHash: faviconhash,
	}, nil
}

// resolveURL 将相对URL转换为完整URL
func ResolveURL(baseURL string, href string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	u, err := url.Parse(href)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return base.ResolveReference(u).String(), nil
}

// 切片中内容去重
func RemoveDuplicates(slice []string) []string {
	// Create a map to track seen elements
	seen := make(map[string]struct{})
	result := []string{}

	for _, item := range slice {
		// If the item is not in the map, add it to the result slice
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
