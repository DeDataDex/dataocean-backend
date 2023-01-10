package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmitPaySign = "submit_pay_sign"

var _ sdk.Msg = &MsgSubmitPaySign{}

func NewMsgSubmitPaySign(creator string, paySign string, payData string) *MsgSubmitPaySign {
	return &MsgSubmitPaySign{
		Creator: creator,
		PaySign: paySign,
		PayData: payData,
	}
}

func (msg *MsgSubmitPaySign) Route() string {
	return RouterKey
}

func (msg *MsgSubmitPaySign) Type() string {
	return TypeMsgSubmitPaySign
}

func (msg *MsgSubmitPaySign) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitPaySign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitPaySign) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
