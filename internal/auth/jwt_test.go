package auth

import (
	"testing"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"time"
	"net/http"
	"fmt"
)


// function signatures:
// func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error)
// func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error)

// utils
var userID = uuid.New()
var token, _ = MakeJWT(userID, "good", 5*time.Minute)
var expiredToken, _ = MakeJWT(userID, "good", -time.Second)

func TestValidateJWT(t *testing.T) {
	
	tests := map[string]struct {
		tokenString string
		tokenSecret string
		want        uuid.UUID
		wantErr     bool
	}{
		"success": {tokenString: token, tokenSecret: "good", want: userID, wantErr: false},
		"expired":  {tokenString: expiredToken,tokenSecret: "good", want: uuid.Nil, wantErr: true},
		"invalid": {tokenString: token, tokenSecret: "bad", want: uuid.Nil, wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotResult, gotErr := ValidateJWT(tc.tokenString, tc.tokenSecret)
			diffResult := cmp.Diff(tc.want, gotResult)
			if diffResult != "" {
				t.Fatalf("%s", diffResult)
			}
			if tc.wantErr && gotErr == nil {
				t.Fatal("expected an error got nil")
			}
			if !tc.wantErr && gotErr != nil {
				t.Fatalf("unexpected error: %v", gotErr)
			}
		})
	}
}


// function signature:
//func GetBearerToken(headers http.Header) (string, error)

func TestGetBearerToken(t *testing.T) {

// utils
req1, _ := http.NewRequest("POST", "http://localhost:8080/dosomething",nil)
req1.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
req2, _ := http.NewRequest("POST", "http://localhost:8080/dosomething",nil)

	tests := map[string]struct {
		headers http.Header
		want    string
		wantErr bool
	}{
		"success": {headers: req1.Header, want: token, wantErr: false},
		"no_header": {headers: req2.Header, want: "", wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotResult, gotErr := GetBearerToken(tc.headers)
			diffResult := cmp.Diff(tc.want, gotResult)
			if diffResult != "" {
				t.Fatalf("%s", diffResult)
			}
			if tc.wantErr && gotErr == nil {
				t.Fatal("expected an error got nil")
			}
			if !tc.wantErr && gotErr != nil {
				t.Fatalf("unexpected error: %v", gotErr)
			}
		})
	}
}
















