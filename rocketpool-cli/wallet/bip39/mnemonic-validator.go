package bip39

import (
	"errors"
	"sort"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

type MnemonicValidator struct {
	mnemonic []string
}

func Create(length int) *MnemonicValidator {

	if length <= 0 {
		return nil
	}

	out := &MnemonicValidator{}

	out.mnemonic = make([]string, 0, length)

	return out
}

func (mv *MnemonicValidator) AddWord(input string) error {
	wordList := bip39.GetWordList()

	idx := sort.SearchStrings(wordList, input)
	if idx >= len(wordList) {
		return errors.New("Invalid word")
	}

	if wordList[idx] != input && (len(input) < 4 || (len(wordList[idx]) < 4 || wordList[idx][:4] != input[:4])) {
		return errors.New("Invalid word")
	}

	mv.mnemonic = append(mv.mnemonic, wordList[idx])
	return nil
}

func (mv *MnemonicValidator) Filled() bool {
	return len(mv.mnemonic) == cap(mv.mnemonic)
}

func (mv *MnemonicValidator) Finalize() (string, error) {

	if !mv.Filled() {
		return "", errors.New("Not enough words were entered.")
	}

	mnemonic := strings.Join(mv.mnemonic, " ")
	if bip39.IsMnemonicValid(mnemonic) {
		return mnemonic, nil
	}

	return "", errors.New("Invalid mnemonic")
}
