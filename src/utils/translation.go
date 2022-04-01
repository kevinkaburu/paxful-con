package utils

import "strings"

var sw = map[string]string{
	"bet_amount_below_minimum": "Samahani, Dau lako halitoshi kufanya ubashiri. Kiwango cha chini cha dau ni Tsh %v. Tafadhali ongeza salio na ujaribu tena. Kwa msaada zaidi tafadhali  tupigie kupitia 0659070700.",
	"bet_amount_above_maximum": "Dau lako limezidi kiwango kinachohitajika. Unaweza kubashiri mpaka kiasi cha Tzs %v pekee kwa dau.",
	"total_balance_not_enough": "Samahani lakini salio lako ya sasa kwenye BETIKA ni Tzs %v bonasi %v, Kuweka ubashiri wako wa Tzs %v, tafadhali ongeza akaunti yako ya Betika.",
	"bet_on_bonus_less_odds":   "Hujafanikiwa kufanya ubashiri. Ili kuweza kutumia BONASI, kiwango cha chini cha odds kinatakiwa kiwe kuanzia %v. Tafadhali pitia chaguzi zako na ujaribu tena.",
	"invalid_game_id":          "GAME: %v/%v  siyo sahihi. Tafadhali ondoa mechi na ujaribu tena.",
}
var en = map[string]string{
	"bet_amount_below_minimum": "Sorry but your bet amount is less that minimum allowed of  %v. Please try again.",
	"bet_amount_above_maximum": "Your stake amount exceeds the maximum allowed. You can only place bets of upto   %v in stake amount.",
	"total_balance_not_enough": "Sorry but your current balance on BETIKA is  %v bonus %v, To place your bet of  %v, please top up your Betika account.",
	"bet_on_bonus_less_odds":   "Sorry cannot create bet, minimum odds accepted for bets on bonus amounts is %v and minimum picks %v.",
	"invalid_game_id":          "Invalid pick %v/%v. Please remove the match and try again.",
}

// GetString
func GetString(lang, stringkey string) string {

	switch strings.ToLower(lang) {
	case "sw":
		return sw[stringkey]
	case "en":
		return en[stringkey]
	default:
		return en[stringkey]

	}

}
