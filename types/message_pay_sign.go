package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgPaySign = "pay_sign"

var _ sdk.Msg = &MsgPaySign{}

func NewMsgPaySign(creator string, videoId uint64, payPublicKey string) *MsgPaySign {
	return &MsgPaySign{
		Creator:      creator,
		VideoId:      videoId,
		PayPublicKey: payPublicKey,
	}
}

func (msg *MsgPaySign) Route() string {
	return RouterKey
}

func (msg *MsgPaySign) Type() string {
	return TypeMsgPaySign
}

func (msg *MsgPaySign) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgPaySign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPaySign) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
