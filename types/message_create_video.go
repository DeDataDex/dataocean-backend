package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateVideo = "create_video"

var _ sdk.Msg = &MsgCreateVideo{}

func NewMsgCreateVideo(creator string, title string, description string, coverLink string, videoLink string, priceMB uint64) *MsgCreateVideo {
	return &MsgCreateVideo{
		Creator:     creator,
		Title:       title,
		Description: description,
		CoverLink:   coverLink,
		VideoLink:   videoLink,
		PriceMB:     priceMB,
	}
}

func (msg *MsgCreateVideo) Route() string {
	return RouterKey
}

func (msg *MsgCreateVideo) Type() string {
	return TypeMsgCreateVideo
}

func (msg *MsgCreateVideo) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateVideo) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateVideo) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
