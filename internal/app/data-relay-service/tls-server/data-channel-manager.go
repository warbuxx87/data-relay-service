package tls_server

import (
	common_message_consts "bitbucket.org/project-alkemi/common-message-pkg/common-message-consts"
	"errors"
)

type DataChannel struct {
	RemoteAddrString   string
	SubscribedChannels []common_message_consts.DataChannelTypeConst
}

type DataChannelManager struct {
	DataChannels []DataChannel
}

func CreateDataChannelManager() *DataChannelManager {

	dataChannelManager := DataChannelManager{}

	return &dataChannelManager
}

func (dataChannelManager *DataChannelManager) AddChannelSubscription(remoteAddr string, dataChannelSubscription common_message_consts.DataChannelTypeConst) error {

	if dataChannelManager.CheckIfChannelSubscribed(remoteAddr, dataChannelSubscription) {
		return errors.New("channel already subscribed to")

	} else {

		for index, element := range dataChannelManager.DataChannels {

			if element.RemoteAddrString == remoteAddr {
				dataChannelManager.DataChannels[index].SubscribedChannels = append(element.SubscribedChannels, dataChannelSubscription)

				return nil
			}
		}

		// If we got here there is no data channel object so we create one
		newDataChannel := DataChannel{RemoteAddrString: remoteAddr}
		newDataChannel.SubscribedChannels = append(newDataChannel.SubscribedChannels, dataChannelSubscription)
		dataChannelManager.DataChannels = append(dataChannelManager.DataChannels, newDataChannel)
	}

	return nil
}

func (dataChannelManager *DataChannelManager) CheckIfChannelSubscribed(remoteAddr string, dataChannelSubscription common_message_consts.DataChannelTypeConst) bool {

	found := false

	for _, element := range dataChannelManager.DataChannels {

		if element.RemoteAddrString == remoteAddr {

			for _, chanElement := range element.SubscribedChannels {

				if chanElement == dataChannelSubscription {
					found = true
				}

			}

		}
	}

	return found
}

func (dataChannelManager *DataChannelManager) RemoveChannelSubscription(remoteAddr string,
	dataChannelSubscription common_message_consts.DataChannelTypeConst) error {

	if dataChannelManager.CheckIfChannelSubscribed(remoteAddr, dataChannelSubscription) == false {
		return errors.New("channel is not subscribed to")

	} else {

		for index, element := range dataChannelManager.DataChannels {

			if element.RemoteAddrString == remoteAddr {
				dataChannelManager.DataChannels[index].SubscribedChannels = append(element.SubscribedChannels[:index], element.SubscribedChannels[index+1:]...)

				return nil
			}
		}

		// If we got here no data channel exist

		return errors.New("data channel does not exist for " + remoteAddr)

	}
}

func (dataChannelManager *DataChannelManager) RemoveDataChannel(remoteAddr string) error {

	for index, element := range dataChannelManager.DataChannels {

		if element.RemoteAddrString == remoteAddr {
			dataChannelManager.DataChannels = append(dataChannelManager.DataChannels[:index], dataChannelManager.DataChannels[index+1:]...)
			return nil
		}
	}

	return errors.New("data channel for " + remoteAddr + " does not exist")
}
