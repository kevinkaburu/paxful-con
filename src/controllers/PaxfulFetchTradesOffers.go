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

	pticker := time.NewTicker(60 * time.Second)
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
		var fiats models.FiatCurency
		if err = rows.Scan(&fiats.FiatCurrencyId, &fiats.FiatCurrencyCode); err != nil {
			log.Printf("unable to read Currency record %v", err)
			continue
		}
		go s.httpOffers(fiats)
	}

}

func (s *Server) httpOffers(fiatCurrency models.FiatCurency) {
	data := url.Values{}
	data.Set("offer_type", "buy")
	data.Set("type", "buy")
	data.Set("limit", "200")
	data.Set("currency_code", fiatCurrency.FiatCurrencyCode)
	data.Set("crypto_currency_code", "USDT")
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
		provider_id := os.Getenv("PAXFUL_EXCHANGE_ID")
		if len(paxfulOffers.Data.Offers) > 0 {
			UpdateOffersStatusQuery := "update offer set status = 0 where exchange_id =? and fiat_currency_id = ?"
			_, err = s.DB.Exec(UpdateOffersStatusQuery, provider_id, fiatCurrency.FiatCurrencyId)
			if err != nil {
				log.Printf("unable to update  to offer  because %v", err)

			}
		}

		for i := 0; i < len(paxfulOffers.Data.Offers); i++ {

			crypto_currency_id := os.Getenv("CRYPTO_USDT_ID")
			offertype := paxfulOffers.Data.Offers[i].OfferType
			external_id := paxfulOffers.Data.Offers[i].OfferID
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
			query := fmt.Sprintf("SELECT payment_type_id from payment_type WHERE name='%v'", payment_method_group)
			err = s.DB.QueryRow(query).Scan(&PaymenttypeID)
			if err != nil {
				continue
			}

			var exchangeID int64
			cquery := "SELECT exchange_id from exchange WHERE name='paxful'"
			err = s.DB.QueryRow(cquery).Scan(&exchangeID)
			if err != nil {
				continue
			}

			insertPaymentQuery := "insert ignore into payment_method (label,payment_type_id	) values (?,?)"
			_, err = s.DB.Exec(insertPaymentQuery, payment_method_name, PaymenttypeID)
			if err != nil {
				log.Printf("unable to insert to payment_method %v because %v", payment_method_name, err)

			}

			insertOffer := "insert  into offer (exchange_id,type,external_id,min_fiat_amount,max_fiat_amount,fiat_currency_id,crypto_currency_id,fiat_price_per_crypto,created,modified, status) VALUES (?,?,?,?,?,?,?,?,now(),now(),1) ON DUPLICATE KEY UPDATE min_fiat_amount=?,max_fiat_amount=?,fiat_price_per_crypto=?, modified = now(), status= 1;"
			OfferObject, err := s.DB.Exec(insertOffer, exchangeID, offertype, external_id, min_fiat_amount, max_fiat_amount, fiatCurrency.FiatCurrencyId, crypto_currency_id, fiat_price_per_crypto, min_fiat_amount, max_fiat_amount, fiat_price_per_crypto)
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
				log.Printf("unable to insert to offer_payment_method %v because %v", tags, err)

			}

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
			var country_id uint64

			query := fmt.Sprintf("SELECT fiat_currency_id,country_id from fiat_currency WHERE currency_code='%v'", code)
			err = s.DB.QueryRow(query).Scan(&fiat_currency_id, &country_id)
			if err != nil {
				continue
			}
			if fiat_currency_id < 1 {
				continue
			}
			usdValue := forexData.Data.Currencies[i].Rate.Usd

			fmt.Println(fmt.Printf("Forex UPDATE: %v | CountryID: %v | USD: %v", code, country_id, usdValue))

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
