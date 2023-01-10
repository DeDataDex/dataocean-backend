package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgPlayVideo = "play_video"

var _ sdk.Msg = &MsgPlayVideo{}

func NewMsgPlayVideo(creator string, videoId uint64) *MsgPlayVideo {
	return &MsgPlayVideo{
		Creator: creator,
		VideoId: videoId,
	}
}

func (msg *MsgPlayVideo) Route() string {
	return RouterKey
}

func (msg *MsgPlayVideo) Type() string {
	return TypeMsgPlayVideo
}

func (msg *MsgPlayVideo) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgPlayVideo) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPlayVideo) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
