// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [4]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
	lider int
}

const ( // tipos de mensagens trocadas entre os processos
	FALHA = iota
	ELEICAO
	LIDER
	TERMINATE 
)

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()
	var temp mensagem
	// comandos para o anel iciam aqui
	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = FALHA
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	chans[3] <- temp

	for true {
		proximoLider := <-in

		if proximoLider == 1 {
			temp.tipo = TERMINATE //quebra o loop dos processos
			chans[(proximoLider+3)%4] <- temp
			break
		}
		fmt.Printf("Controle: confirmação líder escolhido foi %d\n", proximoLider) // receber e imprimir confirmação

		// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)

		fmt.Printf("Controle: mudar o processo %d para falho\n", proximoLider)
		temp.tipo = FALHA
		chans[(proximoLider+3)%4] <- temp
	}

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo

	var actualLeader int
	var bFailed bool = false // todos inciam sem falha

	actualLeader = leader // indicação do ider veio por parâmatro

	Loop:
	for true {

		temp := <-in // ler mensagem
		switch temp.tipo {
		case FALHA:
			{
				bFailed = true
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				var eleicaoMensagem mensagem
				eleicaoMensagem.tipo = ELEICAO     					// indica que é uma mensagem de eleicao
				eleicaoMensagem.corpo[TaskId] = -1 					//indica que está inativo
				out <- eleicaoMensagem
				fmt.Printf("Eleicao em andamento...\n")
				resultadoEleicao := <-in 							// rececbe o corpo da mensagem com todos TaskId's
				proximoLider := -1 									//verifica qual o maior TaskId
				for i := 0; i < 4; i++ {
					if resultadoEleicao.corpo[i] > proximoLider {
						proximoLider = resultadoEleicao.corpo[i]
					}
				}
				fmt.Printf("%d escolheu novo Lider: %d\n", TaskId, proximoLider)
				eleicaoMensagem.tipo = LIDER						 //inidica que é uma mensagem contendo o lider
				eleicaoMensagem.lider = proximoLider
				out <- eleicaoMensagem
			}
		case ELEICAO:
			{
				//bFailed = false
				if bFailed {
					fmt.Printf("%2d: está inativo\n", TaskId)
					temp.corpo[TaskId] = -1
				}else{
					fmt.Printf("%2d: colocou o TaskId\n", TaskId)
					temp.corpo[TaskId] = TaskId;
				}
				out <- temp
			}
		case LIDER:
			{
				if actualLeader == temp.lider { 					//identifica que a mensagem deu uma volta
					controle <- actualLeader						// avisa o controle para falhar o proximo lider
				} else {
					actualLeader = temp.lider
					fmt.Printf("%d: recebeu novo lider: %d\n", TaskId, actualLeader)
					out <- temp
				}
			}
		case TERMINATE:
			{
				if TaskId == actualLeader { 
					fmt.Printf("%2d: Mensagem de erro chegou no lider %d\n", TaskId, actualLeader)
					out <- temp 			//lider manda a mensagem de erro para o proximo 
					<- in					//espera que de a volta para terminar
					break Loop
				}else {
					fmt.Printf("%2d: encerrou\n", TaskId)
					out <- temp 			//termina o proximo processo
					break Loop
				}
			}
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 0) // este é o lider
	go ElectionStage(1, chans[0], chans[1], 0) // não é lider, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // não é lider, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
