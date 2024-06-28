package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AplikasiRentasDigital/eways-enigma-consumer/common"
	commonmaster "github.com/AplikasiRentasDigital/eways-enigma-master/common"
	"github.com/AplikasiRentasDigital/eways-enigma-master/database"
	"github.com/AplikasiRentasDigital/eways-enigma-master/middleware"
	"github.com/spf13/viper"

	//"github.com/AplikasiRentasDigital/eways-enigma-consumer/vendor/github.com/AplikasiRentasDigital/eways-enigma-master/repositories"
	"github.com/AplikasiRentasDigital/eways-enigma-master/logic/api"

	engrepo "github.com/AplikasiRentasDigital/eways-enigma-master/repositories"
	service "github.com/AplikasiRentasDigital/eways-enigma-master/services"
	enigmastruct "github.com/AplikasiRentasDigital/eways-enigma-master/structs"
)

// var (

//	//inboundlogic      api.InboundLogic             = api.NewInboundLogic(inboundservice)
//
// )
var ctx = context.Background()
var RedisUrl string = ""
var RedisUrlPrimary string = ""
var RedisUrlReplica string = ""

func main() {

	middleware.NewViperLoad()
	client := common.InitializeRedis(common.REDIS_DB)
	DB := database.NewMYSQLConn()

	outboundRepository := engrepo.NewOutboundRepository(DB)
	outboundService := service.NewOutboundService(outboundRepository, DB)

	inboundrepository := engrepo.NewInboundRepository(DB)
	inboundservice := service.NewInboundService(inboundrepository, outboundRepository, DB)

	inboundwarepository := engrepo.NewInboundWARepository(DB)
	inboundwaservice := service.NewInboundWAService(inboundwarepository, DB)

	statusreportrepository := engrepo.NewStatusReportRepository(DB)
	statusreportservice := service.NewStatusReportService(statusreportrepository)
	statusreportlogic := api.NewStatusReportLogic(statusreportservice, inboundservice, DB)

	conversationRepository := engrepo.NewConversationRepository(DB)
	conversationService := service.NewConversationService(conversationRepository, outboundRepository, DB)

	logReository := engrepo.NewLogRepository(DB)
	logService := service.NewLogService(logReository)
	isLoop := true

	fmt.Printf("Server running version 0.0.8\n")

	for {
		// do something
		if !isLoop { // the condition stops matching
			break // break out of the loop
		}

		go func() {
			for {

				var outbound enigmastruct.OutboundConsumer
				result, err := client.BLPop(ctx, 0*time.Second, common.OUTBOUND_QUEUE).Result()

				if err != nil {
					fmt.Println(err.Error())
				}
				startAll := time.Now()
				json.Unmarshal([]byte(result[1]), &outbound)

				for _, outboundMessage := range outbound.Messages {

					startSingle := time.Now()
					t := make(map[string]enigmastruct.VendorService)
					for k := range outbound.VendorService {
						//if outbound.VendorService[k].FromSender == outboundMessage.From {
						//	t[outbound.VendorService[k].Channel] = outbound.VendorService[k]
						//}
						t[outbound.VendorService[k].Channel] = outbound.VendorService[k]
					}
					outboundService.SendOutboundSgl(outboundMessage, t, outbound.ClientID)
					elapsedSingle := time.Since(startSingle)
					fmt.Println("Single Outbound Took ", outboundMessage.Channel, outboundMessage.To, elapsedSingle)
				}

				elapsedAll := time.Since(startAll)
				fmt.Println("All Outbound Took ", elapsedAll)

			}
		}()

		go func() {
			for {

				// var outboundLog enigmastruct.OutboundBulkLog
				var outbound enigmastruct.OutboundConsumer
				result, err := client.BLPop(ctx, 0*time.Second, common.OUTBOUND_QUEUE_BULK).Result()

				if err != nil {
					log.Println("error go routine in line 108_main ", err.Error())
				} else {
					startAll := time.Now()
					if len(result) > 0 {
						json.Unmarshal([]byte(result[1]), &outbound)
						// json.Unmarshal([]byte(outboundLog.RequestBody), &outbound)

						var outMessages []*enigmastruct.HTTPRequest
						t := make(map[string]enigmastruct.VendorService)
						for k := range outbound.VendorService {
							t[outbound.VendorService[k].Channel] = outbound.VendorService[k]
						}

						for j, outboundMessage := range outbound.Messages {

							startSingle := time.Now()

							outResp := outboundService.SendOutboundBulk(outboundMessage, t, outbound.ClientID)
							elapsedSingle := time.Since(startSingle)

							outMessages = append(outMessages, &outResp)
							log.Println("Outbound Bulk Took ", outboundMessage.Channel, outboundMessage.To, elapsedSingle)

							if len(outMessages) == 200 || j+1 == len(outbound.Messages) {
								logService.InsertLogBulk(outMessages)
								outboundService.InsertOutboundBulk(outMessages, outbound.ClientID)
								outMessages = []*enigmastruct.HTTPRequest{}
								bulkTime := time.Since(startAll)
								log.Println("Bulk insert outbound and api_log : ", bulkTime)
							}

							// TODO: fix this go func code
							// go func(oM enigmastruct.OutMessage, tData map[string]enigmastruct.VendorService, obnd enigmastruct.OutboundConsumer) {
							// 	startSingle := time.Now()

							// 	outResp := outboundService.SendOutboundBulk(oM, tData, obnd.ClientID)
							// 	elapsedSingle := time.Since(startSingle)

							// 	outMessages = append(outMessages, &outResp)
							// 	log.Println("Outbound Bulk Took ", oM.Channel, oM.To, elapsedSingle)

							// 	if len(outMessages) == 200 || j+1 == len(obnd.Messages) {
							// 		logService.InsertLogBulk(outMessages)
							// 		outboundService.InsertOutboundBulk(outMessages, obnd.ClientID)
							// 		outMessages = []*enigmastruct.HTTPRequest{}
							// 		bulkTime := time.Since(startAll)
							// 		log.Println("Bulk insert outbound and api_log : ", bulkTime)
							// 	}
							// }(outboundMessage, t, outbound)
						}

						// logService.InsertLogBulk(outMessages)

						elapsedAll := time.Since(startAll)
						//fmt.Println("All Outbound Took ", elapsedAll)
						log.Println("All Outbound Bulk Took ", elapsedAll)
					}
				}
			}
		}()

		// Outbound Simulation
		go func() {
			for {

				// var outboundLog enigmastruct.OutboundBulkLog
				var outbound enigmastruct.OutboundConsumer
				result, err := client.BLPop(ctx, 0*time.Second, commonmaster.OUTBOUND_QUEUE_SIMULATION).Result()

				if err != nil {
					//fmt.Println(err.Error())
					log.Println("error go routine in line 159_main ", err.Error())
				} else {
					startAll := time.Now()
					if result != nil && len(result) > 0 {
						json.Unmarshal([]byte(result[1]), &outbound)
						// json.Unmarshal([]byte(outboundLog.RequestBody), &outbound)

						var outMessages []*enigmastruct.HTTPRequest
						t := make(map[string]enigmastruct.VendorService)
						for k := range outbound.VendorService {
							t[outbound.VendorService[k].Channel] = outbound.VendorService[k]
						}

						for j, outboundMessage := range outbound.Messages {
							startSingle := time.Now()

							outboundMessage.Simulation = true

							outResp := outboundService.SendOutboundBulk(outboundMessage, t, outbound.ClientID)
							elapsedSingle := time.Since(startSingle)

							outMessages = append(outMessages, &outResp)
							log.Println("Outbound Simulation Took ", outboundMessage.Channel, outboundMessage.To, elapsedSingle)

							if len(outMessages) == 200 || j+1 == len(outbound.Messages) {
								logService.InsertLogBulk(outMessages)
								outboundService.InsertOutboundBulk(outMessages, outbound.ClientID)
								outMessages = []*enigmastruct.HTTPRequest{}
								bulkTime := time.Since(startAll)
								log.Println("Insert outbound simulation and api_log : ", bulkTime)
							}
						}

						// logService.InsertLogBulk(outMessages)

						elapsedAll := time.Since(startAll)
						//fmt.Println("All Outbound Took ", elapsedAll)
						log.Println("All Outbound Simulation Took ", elapsedAll)
					}
				}
			}
		}()

		go func() {
			for {

				//var damcorpInbound enigmastruct.Damcorp_Inbound_Request_WA
				result, err := client.BLPop(ctx, 0*time.Second, common.DAMCORP_INBOUND_WA).Result()

				if err != nil {
					log.Println("error go routine in line 207 ")
				}

				startAll := time.Now()
				if len(result) > 0 {
					vs, err2 := inboundservice.GetWebhook("WhatsAppDto", "InboundDto", "SesameDto", "InboundDto")
					if err2.Code != 200 {
						log.Println(err.Error())
					} else {
						//damcorpInbound.Webhook = vs
						//json.Unmarshal([]byte(result[1]), &damcorpInbound)
						inboundwaservice.SendInboundWADamcorp([]byte(result[1]), vs)
					}
				}

				elapsedAll := time.Since(startAll)
				fmt.Println("All Inbound damcorp Took ", elapsedAll)

			}
		}()

		/*
			go func() {
				for {

					var outbound enigmastruct.OutboundConsumer
					result, err := client.BLPop(ctx, 0*time.Second, common.OUTBOUND_QUEUE).Result()

					if err != nil {
						fmt.Println(err.Error())
					}
					startAll := time.Now()
					json.Unmarshal([]byte(result[1]), &outbound)
					for _, outboundMessage := range outbound.Messages {
						startSingle := time.Now()
						t := make(map[string]enigmastruct.VendorService)
						for k := range outbound.VendorService {
							t[outbound.VendorService[k].Channel] = outbound.VendorService[k]
						}
						outboundService.SendOutboundSgl(outboundMessage, t, outbound.ClientID)
						elapsedSingle := time.Since(startSingle)
						fmt.Println("Single Outbound Took ", outboundMessage.Channel, outboundMessage.To, elapsedSingle)
					}

					elapsedAll := time.Since(startAll)
					fmt.Println("All Outbound Took ", elapsedAll)

				}
			}()
		*/

		go func() {
			for {

				//var damcorpInbound enigmastruct.Damcorp_Inbound_Request_WA
				result, err := client.BLPop(ctx, 0*time.Second, common.CONVERSATION_QUEUE).Result()

				if err != nil {
					log.Printf("error BLPOP : %s", err.Error())
				}

				startAll := time.Now()
				if len(result) > 0 {
					conversationService.ConversationDevlivery([]byte(result[1]))
					endTime := time.Since(startAll)
					log.Println("Conversation log time process : ", endTime)
				}

				elapsedAll := time.Since(startAll)
				log.Println("All Conversation Whatsapp Took ", elapsedAll)

			}
		}()

		go func() {
			for {
				//var statusreportsparkpost enigmastruct.StatusReportSparkPost
				var statusreportsparkposts []enigmastruct.StatusReportSparkPost
				const SPARKPOST_QUEUE_REPORT string = "sparkpost-queue-report"
				result, err := client.BLPop(ctx, 0*time.Second, SPARKPOST_QUEUE_REPORT).Result()

				if err != nil {
					fmt.Println(err.Error())
				}
				if result != nil {
					startAll := time.Now()
					json.Unmarshal([]byte(result[1]), &statusreportsparkposts)

					startSingle := time.Now()
					bln := statusreportlogic.ProcessStatusSparkPostConsumer(statusreportsparkposts)
					/*
						if bln {
							webhooks, err := inboundservice.GetWebhook("SparkPostStatus", "enigma", "galactic", "SparkPostStatus")
							_ = err
							statusreportsparkpostsByte, _ := json.Marshal(statusreportsparkposts)

							for i := range webhooks {
								_, _, body, _, errhit := common.HitAPI(webhooks[i].URL, statusreportsparkpostsByte, webhooks[i].Method, webhooks[i].Token, time.Duration(0))
								_ = body
								if errhit != nil {
									fmt.Println(errhit)
								}

							}
						}
					*/
					elapsedSingle := time.Since(startSingle)
					fmt.Println("Finish Process Status Sparkpost Took ", "SparkPost ", "with result ", bln, statusreportsparkposts[0].Msys.MessageEvent.RcptTo, elapsedSingle)
					elapsedAll := time.Since(startAll)
					fmt.Println("All Process Status Sparkpost Took ", elapsedAll)
				}
			}

		}()

		select {}
	}

	/*
		client := common.InitializeRedis(common.REDIS_DB)
		go func() {
			for {
				var outbound enigmastruct.OutboundConsumer
				result, err := client.BLPop(ctx, 0*time.Second, common.OUTBOUND_KEYS).Result()

				if err != nil {
					fmt.Println(err.Error())
				}
				startAll := time.Now()
				json.Unmarshal([]byte(result[1]), &outbound)
				for _, outboundMessage := range outbound.Messages {
					startSingle := time.Now()
					t := make(map[string]enigmastruct.VendorService)
					for k := range outbound.VendorService {
						t[outbound.VendorService[k].Channel] = outbound.VendorService[k]
					}
					outboundService.SendOutboundSgl(outboundMessage, t, outbound.ClientID)
					elapsedSingle := time.Since(startSingle)
					fmt.Println("Single Outbound Took ", outboundMessage.Channel, outboundMessage.To, elapsedSingle)
				}

				elapsedAll := time.Since(startAll)
				fmt.Println("All Outbound Took ", elapsedAll)

			}
		}()

		go func() {
			chkLoop := true
			for {
				if !chkLoop { // the condition stops matching
					break // break out of the loop
				}
				for {
					result, err := client.BLPop(ctx, 0*time.Second, common.SPARKPOST_QUEUE_REPORT).Result()
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					key := common.SPARKPOST_QUEUE_REPORT
					var statusreportsparkpost enigmastruct.StatusReportSparkPost
					var statusreportsparkposts []enigmastruct.StatusReportSparkPost
					json.Unmarshal([]byte(result[0]), &statusreportsparkpost)
					statusreportsparkposts = append(statusreportsparkposts, statusreportsparkpost)

					for chkLoop {
						chkLoop = GetRunningRedis(key)
						if !chkLoop {
							break
						}

						//loop again in 3 secs
						time.Sleep(5 * time.Second)
					}
					common.Set(key, statusreportsparkpost, common.REDIS_DB)

					start := time.Now()
					fmt.Println("Start Process Status Sparkpost " + statusreportsparkpost.Msys.MessageEvent.CustomerID)

					statusreportlogic.ProcessStatusSparkPostConsumer(statusreportsparkposts)
					//uniquecodeLogic.GenerateUniqueCode(uniqueCodeRequest)
					common.Delete(key, common.REDIS_DB)
					elapsed := time.Since(start)
					fmt.Println("Finish Process Status Sparkpost "+statusreportsparkpost.Msys.MessageEvent.CustomerID, elapsed)
					chkLoop = true
				}
			}

		}()

		select {}
	*/

	serverPort := viper.Get("PORT").(string)
	if serverPort == "" {
		serverPort = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil))

}

func GetRunningRedis(keyredis string) bool {

	//search key all
	resAll, errAll := common.Get(keyredis, common.REDIS_DB)
	if errAll != nil {
		fmt.Println("error when get the key cust redeem All:", keyredis)
		fmt.Println(errAll.Error())
		return true
	}
	if resAll != "" {
		return true
	}

	return false
}
