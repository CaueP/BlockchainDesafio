/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Implementação iniciada por Caue Garcia Polimanti
blockchain_dojo com identificação do usuário que realizou a request
*/

// nome do package
package main

// lista de imports
import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"	
	"reflect"
	
	"log"
	"bytes"	
	"net/http"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"io/ioutil"
)
// "github.com/op/go-logging"
//var myLogger = logging.MustGetLogger("dojo_mgm")

// BoletoPropostaChaincode - implementacao do chaincode
type BoletoPropostaChaincode struct {
}

// Definição da Struct Proposta e parametros para exportação para JSON
type Proposta struct {
    ID					string	`json:"id_proposta"`
	CpfPagador			string 	`json:"cpf_pagador"`
	PagadorAceitou 		bool 	`json:"pagador_aceitou"`
	BeneficiarioAceitou bool 	`json:"beneficiario_aceitou"`
	BoletoPago 			bool 	`json:"boleto_pago"`
}

// consts associadas à tabela de Propostas
const (
	nomeTabelaProposta		=	"Proposta"
	colCpfPagador			=	"cpfPagador"
	colPagadorAceitou		=	"pagadorAceitou"
	colBeneficiarioAceitou	=	"beneficiarioAceitou"
	colBoletoPago			=	"boletoPago"
)

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(BoletoPropostaChaincode))
	if err != nil {
		fmt.Printf("Error starting BoletoPropostaChaincode chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init
// 		Inicia/Reinicia a tabela de propostas
// ============================================================================================================================
func (t *BoletoPropostaChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//myLogger.Debug("Init Chaincode...")
	fmt.Println("Init Chaincode...")

	// Verificação da quantidade de argumentos recebidos
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	// Verifica se a tabela 'Proposta' existe
	fmt.Println("Verificando se a tabela " + nomeTabelaProposta + " existe...")
	tbProposta, err := stub.GetTable(nomeTabelaProposta)
	if err != nil {
		fmt.Println("Falha ao executar stub.GetTable para a tabela " + nomeTabelaProposta + ". [%v]", err)
	}
	// Se a tabela 'Proposta' já existir, excluir a tabela
	if tbProposta != nil {	
		err = stub.DeleteTable(nomeTabelaProposta)
		fmt.Println("Tabela " + nomeTabelaProposta + " excluída.")
	}


	// Criar tabela de Propostas
	fmt.Println("Criando a tabela " + nomeTabelaProposta + "...")
	err = stub.CreateTable(nomeTabelaProposta, []*shim.ColumnDefinition{
		// Identificador da proposta (hash)
		&shim.ColumnDefinition{Name: "Id", Type: shim.ColumnDefinition_STRING, Key: true},
		// CPF do Pagador
		&shim.ColumnDefinition{Name: colCpfPagador, Type: shim.ColumnDefinition_STRING, Key: false},
		// Status de aceite do Pagador da proposta
		&shim.ColumnDefinition{Name: colPagadorAceitou, Type: shim.ColumnDefinition_BOOL, Key: false},
		// Status de aceite do Beneficiario da proposta
		&shim.ColumnDefinition{Name: colBeneficiarioAceitou, Type: shim.ColumnDefinition_BOOL, Key: false},
		// Status do Pagamento do Boleto
		&shim.ColumnDefinition{Name: colBoletoPago, Type: shim.ColumnDefinition_BOOL, Key: false},
	})
	if err != nil {
		return nil, fmt.Errorf("Falha ao criar a tabela " + nomeTabelaProposta + ". [%v]", err)
	} 
	fmt.Println("Tabela " + nomeTabelaProposta + " criada com sucesso.")



	// Set the admin
	// The metadata will contain the certificate of the administrator
	adminMeta, err := stub.GetCallerMetadata()
	if err != nil {
		fmt.Println("Failed getting metadata")
		//return nil, errors.New("Failed getting metadata.")
	}
	if len(adminMeta) == 0 {
		fmt.Println("Invalid admin certificate (adminMeta). Empty.")
		//return nil, errors.New("Invalid admin certificate (adminMeta). Empty.")
	}

	fmt.Println("The administrator is (adminMeta) [%x]", adminMeta)
/*
	adminCert, err := stub.GetCallerCertificate()
	if err != nil {
		fmt.Println("Failed getting metadata")
		//return nil, errors.New("Failed getting metadata.")
	}
	if len(adminCert) == 0 {
		fmt.Println("Invalid admin certificate (adminCert). Empty.")
		//return nil, errors.New("Invalid admin certificate (adminCert). Empty.")
	}

	fmt.Println("The administrator is (adminCert) [%x]", adminCert)
*/
	stub.PutState("admin", adminMeta)


	fmt.Println("Init Chaincode... Finalizado!")

	return nil, nil
}


// ============================================================================================================================
// Invoke Functions
// ============================================================================================================================

// Invoke - Ponto de entrada para chamadas do tipo Invoke.
// Funções suportadas:
// "init": inicializa o estado do chaincode, também utilizado como reset
// "registrarProposta(Id, cpfPagador, pagadorAceitou, 
// beneficiarioAceitou, boletoPago)": para registrar uma nova proposta ou atualizar uma já existente.
// Only an administrator can call this function.
// "consultarProposta(Id)": para consultar uma Proposta existente. 
// Only the owner of the specific asset can call this function.
// An asset is any string to identify it. An owner is representated by one of his ECert/TCert.
func (t *BoletoPropostaChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//myLogger.Debug("Invoke Chaincode...")
	fmt.Println("Invoke Chaincode...")
	fmt.Println("invoke is running " + function)

	// Estrutura de Seleção para escolher qual função será chamada, 
	// de acordo com a funcao chamada
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "registrarProposta" {
		return t.registrarProposta(stub, args)
	}
	fmt.Println("invoke não encontrou a func: " + function) //error

	return nil, errors.New("Invocação de função desconhecida: " + function)
}

// registrarProposta: função Invoke para registrar uma nova proposta, recebendo os seguintes argumentos:
// args[0]: Id. Hash que identificará a proposta
// args[1]: cpfPagador. CPF do Pagador
// args[2]: pagadorAceitou. Status de aceite do Pagador da proposta
// args[3]: beneficiarioAceitou. Status de aceite do Beneficiario da proposta
// args[4]: boletoPago. Status do Pagamento do Boleto
func (t *BoletoPropostaChaincode) registrarProposta(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//myLogger.Debug("registrarProposta...")
	fmt.Println("registrarProposta...")

	var jsonResp string

	// Verifica se a quantidade de argumentos recebidas corresponde a esperada
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

	// Obtem os valores da array de arguments (args) e 
	// os converte no tipo necessário para salvar na tabela 'Proposta'
	idProposta := args[0]
	cpfPagador := args[1]
	pagadorAceitou, err := strconv.ParseBool(args[2])
	if err != nil {
		return nil, errors.New("Failed decodinf pagadorAceitou")
	}
	beneficiarioAceitou, err := strconv.ParseBool(args[3])
	if err != nil {
		return nil, errors.New("Failed decodinf beneficiarioAceitou")
	}
	boletoPago, err := strconv.ParseBool(args[4])
	if err != nil {
		return nil, errors.New("Failed decodinf boletoPago")
	}

	// [To do] verificar identidade

	// Verify the identity of the caller
	// Only an administrator can invoker assign
	adminCertificate, err := stub.GetState("admin")
	if err != nil {
		return nil, errors.New("Failed fetching admin identity")
	}

	ok, err := t.isCaller(stub, adminCertificate)
	if err != nil {
		return nil, errors.New("Failed checking admin identity")
	}
	if !ok {
		return nil, errors.New("The caller is not an administrator")
	}

	// Registra a proposta na tabela 'Proposta'
	fmt.Println("Criando Proposta Id [" + idProposta + "] para CPF nº ["+ cpfPagador +"]")
	fmt.Printf("pagadorAceitou: " + strconv.FormatBool(pagadorAceitou)) 
	fmt.Printf(" | beneficiarioAceitou: " + strconv.FormatBool(beneficiarioAceitou))
	fmt.Printf(" | boletoPago: " + strconv.FormatBool(boletoPago) + "\n")

	ok, err = stub.InsertRow(nomeTabelaProposta, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: idProposta}},
			&shim.Column{Value: &shim.Column_String_{String_: cpfPagador}},
			&shim.Column{Value: &shim.Column_Bool{Bool: pagadorAceitou}},
			&shim.Column{Value: &shim.Column_Bool{Bool: beneficiarioAceitou}},
			&shim.Column{Value: &shim.Column_Bool{Bool: boletoPago}} },
	})

	// Caso a proposta já exista (false and no error if a row already exists for the given key).
	if !ok && err == nil {
		// Apenas retornar que a proposta existe
		//return nil, errors.New("Proposta já existente.")
		//jsonResp = "{\"registrado\":\"" + "False" + "\"}"
		//return []byte(jsonResp), errors.New("Proposta já existente.")

		// /* 
		// Trecho para atualizar uma proposta existente
		//	substitui um registro existente em uma linha com o registro associado ao idProposta recebido nos argumentos
		ok, err := stub.ReplaceRow(nomeTabelaProposta, shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: idProposta}},
			&shim.Column{Value: &shim.Column_String_{String_: cpfPagador}},
			&shim.Column{Value: &shim.Column_Bool{Bool: pagadorAceitou}},
			&shim.Column{Value: &shim.Column_Bool{Bool: beneficiarioAceitou}},
			&shim.Column{Value: &shim.Column_Bool{Bool: boletoPago}} },
		})

		

		if !ok && err == nil {
			return nil, errors.New("Falha ao atualizar a Proposta nº " + idProposta)
		}

		


		// chamada Rest API externa
		fmt.Println("Iniciando chamada de API externa")

		// Proposta a ser enviada
		var resProposta Proposta

		// Criação do objeto Proposta	
		resProposta.ID = idProposta
		resProposta.CpfPagador = cpfPagador
		resProposta.PagadorAceitou = pagadorAceitou
		resProposta.BeneficiarioAceitou = beneficiarioAceitou
		resProposta.BoletoPago = boletoPago

		// Converte a proposta em json
		jsonStr, err := json.Marshal(resProposta)

		// define a URL do POST
		url := "http://bc-desafio.mybluemix.net/atualizar"
		fmt.Println("URL:>", url)

		// Build the request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Fatal("NewRequest: ", err)
		}
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")


		// For control over HTTP client headers,
		// redirect policy, and other settings,
		// create a Client
		// A Client is an HTTP client
		client := &http.Client{}


		// Send the request via a client
		// Do sends an HTTP request and
		// returns an HTTP response
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		
		// Callers should close resp.Body
		// when done reading from it
		// Defer the closing of the body
		defer resp.Body.Close()

		// logs de resposta
		fmt.Println("Logs de Resposta")
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))

		return nil, nil
		//*/
	}

	//myLogger.Debug("Proposta criada!")
	fmt.Println("Proposta criada!")

	jsonResp = "{\"registrado\":\"" + "true" + "\"}"
	return []byte(jsonResp), err
}


// ============================================================================================================================
// Query
// ============================================================================================================================

// Query is our entry point for queries

// Query - Ponto de entrada para chamadas do tipo Query.
// Funções suportadas:
// "consultarProposta(Id)": para consultar uma proposta existente
func (t *BoletoPropostaChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//myLogger.Debug("Query Chaincode...")
	fmt.Println("Query Chaincode...")

	fmt.Println("query is running " + function)

	// Estrutura de Seleção para escolher qual função será chamada, 
	// de acordo com a funcao chamada
	if function == "consultarProposta" { //read a variable
		// Consultar uma Proposta existente
		return t.consultarProposta(stub, args)
	}
	fmt.Println("query encontrou a func: " + function) //error

	return nil, errors.New("Query de função desconhecida: " + function)
}

// consultarProposta: função Query para consultar uma proposta existente, recebendo os seguintes argumentos
// args[0]: Id. Hash da proposta
func (t *BoletoPropostaChaincode) consultarProposta(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	//myLogger.Debug("consultarProposta...")
	fmt.Println("consultarProposta...")
	//var listaPropostas []Proposta	// lista de Propostas
	var resProposta Proposta		// Proposta
	var propostaAsBytes []byte			// retorno do json em bytes
	
	// Verifica se a quantidade de argumentos recebidas corresponde a esperada
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Obtem os valores dos argumentos e os prepara para salvar na tabela 'Proposta'
	idProposta := args[0]

	// [To do] verificar identidade

	// Define o valor de coluna do registro a ser buscado
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: idProposta}}
	columns = append(columns, col1)

	// Consultar a proposta na tabela 'Proposta'
	row, err := stub.GetRow(nomeTabelaProposta, columns)
	if err != nil {
		fmt.Println("Erro ao obter Proposta [%s]: [%s]", string(idProposta), err)
		return nil, fmt.Errorf("Erro ao obter Proposta [%s]: [%s]", string(idProposta), err)
	}

	// Tratamento para o caso de não encontrar nenhuma proposta correspondente
	if len(row.Columns) == 0 || row.Columns[2] == nil { 
		return nil, fmt.Errorf("Proposta [%s] não existente.", string(idProposta))	// retorno do erro para o json
	}

	fmt.Println("Query finalizada [% x]", row.Columns[1].GetBytes())

	// Criação do objeto Proposta	
	resProposta.ID = row.Columns[0].GetString_()
	resProposta.CpfPagador = row.Columns[1].GetString_()
	resProposta.PagadorAceitou = row.Columns[2].GetBool()
	resProposta.BeneficiarioAceitou = row.Columns[3].GetBool()
	resProposta.BoletoPago = row.Columns[4].GetBool()

	// Inserir resultado na lista de propostas
	//listaPropostas = append(listaPropostas, resProposta)

	fmt.Println("Proposta: [%s], [%s], [%b], [%b], [%b]", resProposta.ID, resProposta.CpfPagador, resProposta.PagadorAceitou, resProposta.BeneficiarioAceitou, resProposta.BoletoPago)

	// Converter o objeto da Proposta para Bytes, para retorná-lo em formato JSON
	propostaAsBytes, err = json.Marshal(resProposta)
	if err != nil {
			return nil, fmt.Errorf("Query operation failed. Error marshaling JSON: %s", err)
	}
	// retorna o objeto em bytes
	return propostaAsBytes, nil
}


// isCaller: função utilizada para verificar quem é o caller da chamada
func (t *BoletoPropostaChaincode) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	fmt.Println("Check caller...")

	// In order to enforce access control, we require that the
	// metadata contains the signature under the signing key corresponding
	// to the verification key inside certificate of
	// the payload of the transaction (namely, function name and args) and
	// the transaction binding (to avoid copying attacks)

	// Verify \sigma=Sign(certificate.sk, tx.Payload||tx.Binding) against certificate.vk
	// \sigma is in the metadata

	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed getting metadata")
	}/*
	payload, err := stub.GetPayload()
	if err != nil {
		return false, errors.New("Failed getting payload")
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed getting binding")
	}*/

	fmt.Println("passed certificate [% x]", certificate)
	fmt.Println("passed sigma [% x]", sigma)
	//fmt.Println("passed payload [% x]", payload)
	//fmt.Println("passed binding [% x]", binding)

	// valida se os slices são iguais
	if !reflect.DeepEqual(certificate, sigma) {
		fmt.Println("Invalid signature")
		return false, errors.New("Certificado inválido")
	}	

	/*
	ok, err := stub.VerifySignature(
		certificate,
		sigma,
		append(payload, binding...),
	)
	if err != nil {
		fmt.Println("Failed checking signature [%s]", err)
		return ok, err
	} 
	if !ok {
		fmt.Println("Invalid signature")
	}*/

	fmt.Println("Check caller...Verified!")
	// Certificado válido
	return true, nil
	//return ok, err
}