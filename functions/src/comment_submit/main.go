package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/eternal-flame-ad/yumechi.jp/functions/src/comment_submit/comment"
)

func getRequestHeader(key string, request events.APIGatewayProxyRequest) (res []string) {
	for reqHdrKey, val := range request.MultiValueHeaders {
		if strings.ToLower(reqHdrKey) == strings.ToLower(key) {
			res = append(res, val...)
		}
	}
	return
}

func getRequestOrigin(request events.APIGatewayProxyRequest) (string, string, error) {
	origin := getRequestHeader("origin", request)
	if len(origin) != 1 {
		return "", "", errors.New("none or more than one origin headers are present")
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return "", "", errors.New("origin header is not a valid URL")
	}
	return strings.ToLower(u.Scheme), strings.ToLower(u.Host), nil
}

func isAllowedOrigin(request events.APIGatewayProxyRequest) error {
	scheme, host, err := getRequestOrigin(request)
	if err != nil {
		return err
	}
	if scheme == "https" {
		for _, allowedHost := range []string{
			"yumechi.jp",
			"yumechi.jp:443",
			"dev.yumechi.jp",
			"dev.yumechi.jp:443",
			"dev--youthful-pare-0d9fb4.netlify.app",
			"dev--youthful-pare-0d9fb4.netlify.app:443",
		} {
			if host == allowedHost {
				return nil
			}
		}
	}
	return fmt.Errorf("origin (scheme=%s,host=%s) is not acceptable", scheme, host)
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case "OPTIONS":
		origin := getRequestHeader("origin", request)
		if err := isAllowedOrigin(request); err == nil {
			return &events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":  origin[0],
					"Access-Control-Allow-Methods": "POST, OPTIONS",
				},
			}, nil
		} else {
			return nil, err
		}
	case "POST":
		if res := getRequestHeader("content-type", request); len(res) != 1 {
			return nil, fmt.Errorf("unexpected number of content type headers")
		} else if res[0] != "application/x-www-form-urlencoded" {
			return nil, fmt.Errorf("content type %s is not supported", res[0])
		}
		values, err := url.ParseQuery(request.Body)
		if err != nil {
			return nil, err
		}
		_, origin, _ := getRequestOrigin(request)
		formResponse, isClientError, err := comment.HandleForm(values, origin)
		var responseJSON []byte
		if err == nil {
			responseJSON, err = json.Marshal(formResponse)
		}
		if err != nil {
			resp := &events.APIGatewayProxyResponse{
				Body: err.Error(),
			}
			if isClientError {
				resp.StatusCode = http.StatusBadRequest
			} else {
				resp.StatusCode = http.StatusInternalServerError
			}
			return resp, nil
		} else {
			return &events.APIGatewayProxyResponse{
				Body:            comment.Base64Bytes(responseJSON),
				IsBase64Encoded: true,
				StatusCode:      200,
				Headers: map[string]string{
					"content-type": "application/json",
				},
			}, nil
		}
	default:
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       "Only POST requests are allowed",
		}, nil
	}
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}
