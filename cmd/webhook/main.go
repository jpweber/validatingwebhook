package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"io/ioutil"

	"encoding/json"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

func slack() []byte {
	// DEBUG
	log.Println("slack method")
	payload := Payload{
		Text:      "timmy is using latest again",
		Username:  "Milton",
		Channel:   "#hackaton",
		IconEmoji: ":ghost:",
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error marshalling json for slack")
	}
	return jsonBody
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got request...")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		payload, err := json.Marshal(&v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		})
		if err != nil {
			fmt.Println(err)
		}
		w.Write(payload)
	}

	admitResponse := &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: true,
		},
	}

	if ar.Request.Kind.Kind == "Pod" {
		pod := v1.Pod{}
		json.Unmarshal(ar.Request.Object.Raw, &pod)
		for _, container := range pod.Spec.Containers {

			// shame logic here
			// image tag blank or latest
			// linevess or readiness probe empty
			// limits andor requests empty
			tag := strings.Split(container.Image, ":")[1]
			if tag == "latest" || tag == "" {
				fmt.Println("BLOCK container from running...")
				admitResponse.Response.Result = &metav1.Status{
					Message: "No, no, no. you can't use latest",
				}
				// slack call
				body := slack()

				slackURL := "https://hooks.slack.com/services/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"

				_, err := http.Post(slackURL, "json", bytes.NewBuffer(body))
				if err != nil {

					fmt.Print("Error contacting slack")
				}
				break
			} else {
				fmt.Println("Container is a-ok!")
				// send slack notification with gif
			}
		}

	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	payload, err := json.Marshal(admitResponse)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(payload)
}

func main() {
	fmt.Println("webhook starting up...")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServeTLS(":9443", "/certs/server.crt", "/certs/server.key", nil))
}
