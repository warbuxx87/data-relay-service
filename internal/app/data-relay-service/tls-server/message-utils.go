package tls_server

import (
	common_message_consts "bitbucket.org/project-alkemi/common-message-pkg/common-message-consts"
	common_message_models "bitbucket.org/project-alkemi/common-message-pkg/common-message-models"
	"bytes"
	"encoding/gob"
	"github.com/sirupsen/logrus"
	"github.com/getsentry/sentry-go"
	"time"
)

func CreateRequestSuccessMessage(successUUID string, successMessage string) *common_message_models.RawMessageModel {

	var requestSuccessMessageModelBytes bytes.Buffer

	requestSuccessMessageModel := common_message_models.RequestSuccessMessageModel{}

	requestSuccessMessageModel.SuccessUUID = successUUID
	requestSuccessMessageModel.SuccessMessage = successMessage

	requestSuccessMessageMarshalError := gob.NewEncoder(&requestSuccessMessageModelBytes).Encode(requestSuccessMessageModel)

	if requestSuccessMessageMarshalError != nil {
		logrus.Error(requestSuccessMessageMarshalError)
		sentry.CaptureException(requestSuccessMessageMarshalError)
		return nil
	}

	rawMessageModel := common_message_models.RawMessageModel{
		MessageType:     common_message_consts.MessageType_RequestSuccessMessage,
		RawMessageBytes: requestSuccessMessageModelBytes.Bytes(),
		TimeStamp:       time.Now().UTC(),
	}

	return &rawMessageModel
}

func CreateRequestFailureMessage(failedUUID string, errorMessage string) *common_message_models.RawMessageModel {

	var requestFailureMessageModelBytes bytes.Buffer

	requestFailureMessageModel := common_message_models.RequestFailureMessageModel{}

	requestFailureMessageModel.FailedUUID = failedUUID
	requestFailureMessageModel.ErrorMessage = errorMessage

	requestFailureMessageMarshalError := gob.NewEncoder(&requestFailureMessageModelBytes).Encode(requestFailureMessageModel)

	if requestFailureMessageMarshalError != nil {
		logrus.Error(requestFailureMessageMarshalError)
		sentry.CaptureException(requestFailureMessageMarshalError)
		return nil
	}

	rawMessageModel := common_message_models.RawMessageModel{
		MessageType:     common_message_consts.MessageType_RequestFailureMessage,
		RawMessageBytes: requestFailureMessageModelBytes.Bytes(),
		TimeStamp:       time.Now().UTC(),
	}

	return &rawMessageModel
}
