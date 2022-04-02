package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"paxful/src/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func (s *Server) PaxfulFetchOffers() {
	// Fetch active offers from paxful

	pticker := time.NewTicker(2 * time.Minute)
	pquit := make(chan struct{})
	func() {
		for {
			select {
			case <-pticker.C:
				s.FetchActiveOffers()

			case <-pquit:
				pticker.Stop()
				return
			}
		}
	}()
	//fetch Forex httpforex()

	forexTicker := time.NewTicker(30 * time.Minute)
	forexquit := make(chan struct{})
	func() {
		for {
			select {
			case <-forexTicker.C:
				s.httpforex()

			case <-forexquit:
				forexTicker.Stop()
				return
			}
		}
	}()

}

func (s *Server) FetchActiveOffers() {
	fmt.Println("Fetching Offers.... ")

	//fetch possible currencies from db

	activeFiatsQuery := "select fiat_currency_id,currency_code from fiat_currency where status=1"
	rows, err := s.DB.Query(activeFiatsQuery)
	if err != nil {
		fmt.Println(fmt.Printf("unable to fetch fiat Currencies: %v| error: %v", activeFiatsQuery, err))
		return
	}

	for rows.Next() {
		var fiat models.FiatCurency
		if err = rows.Scan(&fiat.FiatCurrencyId, &fiat.FiatCurrencyCode); err != nil {
			log.Printf("unable to read Currency record %v", err)
			continue
		}
		s.httpOffers(fiat, "buy", 0)
		time.Sleep(20 * time.Second)
		s.httpOffers(fiat, "sell", 0)
	}

}

func (s *Server) httpOffers(fiatCurrency models.FiatCurency, offer_type string, offset int) {
	data := url.Values{}
	data.Set("offer_type", offer_type)
	data.Set("type", offer_type)
	data.Set("limit", "300")
	data.Set("offset", fmt.Sprintf("%v", offset))
	data.Set("currency_code", fiatCurrency.FiatCurrencyCode)
	endpoint := fmt.Sprintf("%s/offer/all", os.Getenv("PAXFUL_BASE_URL"))
	//http request
	resp, err := s.PaxfulClient.PostForm(endpoint, data)
	if err != nil {

		log.Printf("error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Printf("error: %v", err)

	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 202 {
		var paxfulOffers models.PaxfulOffers

		if err = json.Unmarshal(body, &paxfulOffers); err != nil {
			fmt.Print("Unable to read response into struct because ", err)
		}
		provider_id := os.Getenv("PAXFUL_PROVIDER_ID")
		if len(paxfulOffers.Data.Offers) > 0 {
			UpdateOffersStatusQuery := "update offer o inner join profile p using(profile_id) set o.status = 0  where p.provider_id =? and o.fiat_currency_id = ?"
			_, err = s.DB.Exec(UpdateOffersStatusQuery, provider_id, fiatCurrency.FiatCurrencyId)
			if err != nil {
				log.Printf("unable to update  to offer  because %v", err)

			}
		}

		for i := 0; i < len(paxfulOffers.Data.Offers); i++ {
			crypto_currency_id := 0
			query := fmt.Sprintf("SELECT crypto_currency_id from crypto_currency WHERE code='%v'", paxfulOffers.Data.Offers[i].CryptoCurrencyCode)
			err = s.DB.QueryRow(query).Scan(&crypto_currency_id)
			if err != nil {
				continue
			}
			offer_owner_username := paxfulOffers.Data.Offers[i].OfferOwnerUsername
			offer_owner_profile_link := paxfulOffers.Data.Offers[i].OfferOwnerProfileLink
			offer_owner_feedback_positive := paxfulOffers.Data.Offers[i].OfferOwnerFeedbackPositive
			offer_owner_feedback_negative := paxfulOffers.Data.Offers[i].OfferOwnerFeedbackNegative
			var profileID int64
			positive := 0
			negative := 0

			profilequery := fmt.Sprintf("SELECT profile_id,positive_feedback,negative_feedback from profile WHERE provider_id=%s and nickname='%v'", provider_id, offer_owner_username)
			err = s.DB.QueryRow(profilequery).Scan(&profileID, &positive, &negative)
			if err != nil {
				insertProfile := "insert  into profile (provider_id,nickname,positive_feedback,negative_feedback,link,created,modified) VALUES (?,?,?,?,?,now(),now())"
				ProfileObject, err := s.DB.Exec(insertProfile, provider_id, offer_owner_username, offer_owner_feedback_positive, offer_owner_feedback_negative, offer_owner_profile_link)
				if err != nil {
					log.Printf("unable to insert to profile to %v because %v", offer_owner_username, err)
					continue
				}
				profileID, _ = ProfileObject.LastInsertId()

			}

			if positive != offer_owner_feedback_positive || negative != offer_owner_feedback_negative {
				updateProfile := "update profile set positive_feedback = ?,negative_feedback=?, modified = now() where profile_id = ?"
				_, err := s.DB.Exec(updateProfile, offer_owner_feedback_positive, offer_owner_feedback_negative, profileID)
				if err != nil {
					log.Printf("unable to update to profile to %v because %v", offer_owner_username, err)
					continue
				}
			}

			external_id := paxfulOffers.Data.Offers[i].OfferID
			OfferLink := paxfulOffers.Data.Offers[i].OfferLink
			min_fiat_amount := paxfulOffers.Data.Offers[i].FiatAmountRangeMin
			max_fiat_amount := paxfulOffers.Data.Offers[i].FiatAmountRangeMax
			fiat_price_per_crypto := paxfulOffers.Data.Offers[i].FiatPricePerCrypto
			payment_method_group := paxfulOffers.Data.Offers[i].PaymentMethodGroup
			payment_method_name := paxfulOffers.Data.Offers[i].PaymentMethodName
			lastSeen := paxfulOffers.Data.Offers[i].LastSeenTimestamp
			insertPaymentGroupQuery := "insert ignore into payment_type (name) values (?)"
			_, err := s.DB.Exec(insertPaymentGroupQuery, payment_method_group)
			if err != nil {
				log.Printf("unable to insert to payment_type %v because %v", payment_method_group, err)

			}
			var PaymenttypeID int64
			queryPtype := fmt.Sprintf("SELECT payment_type_id from payment_type WHERE name='%v'", payment_method_group)
			err = s.DB.QueryRow(queryPtype).Scan(&PaymenttypeID)
			if err != nil {
				continue
			}
			insertPaymentQuery := "insert ignore into payment_method (label,payment_type_id	) values (?,?)"
			_, err = s.DB.Exec(insertPaymentQuery, payment_method_name, PaymenttypeID)
			if err != nil {
				log.Printf("unable to insert to payment_method %v because %v", payment_method_name, err)

			}
			var CountryID int64
			if len(paxfulOffers.Data.Offers[i].CountryName) > 1 {
				queryCoutry := fmt.Sprintf("SELECT country_id from country WHERE name='%v'", paxfulOffers.Data.Offers[i].CountryName)
				_ = s.DB.QueryRow(queryCoutry).Scan(&CountryID)

			}
			OfferType := os.Getenv(offer_type)

			insertOffer := "insert  into offer (profile_id,offer_type_id,country_id,fiat_currency_id,crypto_currency_id,fiat_price_per_crypto,min_fiat_amount,max_fiat_amount,external_link,external_id,status,created,modified) VALUES (?,?,?,?,?,?,?,?,?,?,1,now(),now()) ON DUPLICATE KEY UPDATE min_fiat_amount=?,max_fiat_amount=?,fiat_price_per_crypto=?, modified = now(), status= 1;"
			OfferObject, err := s.DB.Exec(insertOffer, profileID, OfferType, CountryID, fiatCurrency.FiatCurrencyId, crypto_currency_id, fiat_price_per_crypto, min_fiat_amount, max_fiat_amount, OfferLink, external_id, min_fiat_amount, max_fiat_amount, fiat_price_per_crypto)
			if err != nil {
				log.Printf("unable to insert to db to %v because %v", paxfulOffers.Data.Offers[i].OfferID, err)
				return
			}

			var PaymentMethodID int64
			selectquery := fmt.Sprintf("SELECT payment_method_id from payment_method WHERE label='%v' and payment_type_id='%v'", payment_method_name, PaymenttypeID)
			err = s.DB.QueryRow(selectquery).Scan(&PaymentMethodID)
			if err != nil {
				continue
			}
			tags := ""
			for j := 0; j < len(paxfulOffers.Data.Offers[i].Tags); j++ {
				tags = fmt.Sprintf("%s\n%s", tags, paxfulOffers.Data.Offers[i].Tags[j].Description)
			}

			OfferID, err := OfferObject.LastInsertId()
			insertOfferPaymentQuery := "insert ignore into offer_payment_method (offer_id,payment_method_id,tags) values(?,?,?)"
			_, err = s.DB.Exec(insertOfferPaymentQuery, OfferID, PaymentMethodID, tags)
			if err != nil {
				log.Printf("unable to insert to offer_payment_method %v because %v", tags, err)

			}

			insertOfferLastSeen := "insert into offer_last_seen (offer_id,last_seen) values(?,?) ON DUPLICATE KEY UPDATE last_seen=?"
			_, err = s.DB.Exec(insertOfferLastSeen, OfferID, lastSeen, lastSeen)
			if err != nil {
				log.Printf("unable to insert to offer_payment_method %v because %v", lastSeen, err)

			}

		}
		if paxfulOffers.Data.Totalcount > (offset + 300) {
			offset = +300
			go s.httpOffers(fiatCurrency, offer_type, offset)

		}

	} else {
		fmt.Print("Unable to fetch offers", string(body))
		err = errors.New("Error fetching offers")
	}
}

func (s *Server) httpforex() {
	//Get token
	fmt.Println("Fetching forex.... ")

	data := url.Values{}

	endpoint := fmt.Sprintf("%s/currency/list", os.Getenv("PAXFUL_BASE_URL"))
	//http request
	resp, err := s.PaxfulClient.PostForm(endpoint, data)
	if err != nil {
		log.Printf("error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: %v", err)

	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 202 {
		var forexData models.ForexExchange

		if err = json.Unmarshal(body, &forexData); err != nil {
			fmt.Print("Unable to read response into struct because ", err)

		}
		for i := 0; i < len(forexData.Data.Currencies); i++ {

			code := forexData.Data.Currencies[i].Code
			var fiat_currency_id uint64
			query := fmt.Sprintf("SELECT fiat_currency_id from fiat_currency WHERE currency_code='%v'", code)
			_ = s.DB.QueryRow(query).Scan(&fiat_currency_id)
			if fiat_currency_id < 1 {
				continue
			}
			usdValue := forexData.Data.Currencies[i].Rate.Usd

			fmt.Println(fmt.Printf("Forex UPDATE: %v | CurrencyID: %v | USD: %v", code, fiat_currency_id, usdValue))

			insertForex := "insert  into forex_exchange (fiat_currency_id,usd_exchange,created,modified) VALUES (?,?,now(),now()) ON DUPLICATE KEY UPDATE usd_exchange=?, modified = now();"
			if _, err := s.DB.Exec(insertForex, fiat_currency_id, usdValue, usdValue); err != nil {
				log.Printf("unable to insert to forex CURRENCY: %v because %v", code, err)
				return
			}

		}

	} else {
		fmt.Println("Unable to fetch Forex", string(body))
		err = errors.New("Error fetching Forex")
	}
}
