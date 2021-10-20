package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/RtcTokenBuilder"
)

type rtc_int_token_struct struct {
	Uid_rtc_int  uint32 `json:"uid"`
	Channel_name string `json:"ChannelName"`
	Role         uint32 `json:"role"`
}

var rtc_token string
var int_uid uint32
var channel_name string

var role_num uint32
var role rtctokenbuilder.Role

func generateRtcToken(int_uid uint32, channelName string, role rtctokenbuilder.Role) {

	appID := "2997bf2437a74c5489878c5ec224b34d"

	appCertificate := "0720a23244414748a082776246c86b5a"

	expireTimeInSeconds := uint32(40)

	currentTimestamp := uint32(time.Now().UTC().Unix())

	expireTimestamp := currentTimestamp + expireTimeInSeconds

	result, err := rtctokenbuilder.BuildTokenWithUID(appID, appCertificate, channelName, int_uid, role, expireTimestamp)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Token with uid: %s\n", result)
		fmt.Printf("uid is %d\n", int_uid)
		fmt.Printf("ChannelName is %s\n", channelName)
		fmt.Printf("Role is %d\n", role)
	}
	rtc_token = result
}

func rtcTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" && r.Method != "OPTIONS" {
		http.Error(w, "Unsupported method.  Please check.", http.StatusNotFound)
		return
	}

	var t_int rtc_int_token_struct
	var unmarshalErr *json.UnmarshalTypeError
	int_decoder := json.NewDecoder(r.Body)
	int_err := int_decoder.Decode(&t_int)
	if int_err == nil {

		int_uid = t_int.Uid_rtc_int
		channel_name = t_int.Channel_name
		role_num = t_int.Role
		switch role_num {
		case 0:
			// 已废弃。RoleAttendee 和 RolePublisher 的权限相同。
			role = rtctokenbuilder.RoleAttendee
		case 1:
			role = rtctokenbuilder.RolePublisher
		case 2:
			role = rtctokenbuilder.RoleSubscriber
		case 101:
			// 已废弃。RoleAdmin 和 RolePublisher 的权限相同。
			role = rtctokenbuilder.RoleAdmin
		}
	}
	if int_err != nil {

		if errors.As(int_err, &unmarshalErr) {
			errorResponse(w, "Bad request.  Wrong type provided for field "+unmarshalErr.Value+unmarshalErr.Field+unmarshalErr.Struct, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad request.", http.StatusBadRequest)
		}
		return
	}

	generateRtcToken(int_uid, channel_name, role)
	errorResponse(w, rtc_token, http.StatusOK)
	log.Println(w, r)
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["token"] = message
	resp["code"] = strconv.Itoa(httpStatusCode)
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)

}

func main() {

	port := "8082"
	if v := os.Getenv("PORT"); len(v) > 0 {
		port = v
	}

	// 使用 int 型 uid 生成 RTC Token
	http.HandleFunc("/fetch_rtc_token", rtcTokenHandler)
	fmt.Printf("Starting server at port 8082\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
