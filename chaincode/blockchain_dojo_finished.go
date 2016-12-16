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
Implementação iniciada por Caue Garcia Polimanti e Vitor Diego dos Santos de Sousa
*/

// nome do package
package main

// lista de imports
import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"	
)

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
	idProposta				=	"id"
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
		&shim.ColumnDefinition{Name: idProposta, Type: shim.ColumnDefinition_STRING, Key: true},
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

	fmt.Println("Init Chaincode... Finalizado!")

	return nil, nil
}



// ============================================================================================================================
// Invoke Functions
// ============================================================================================================================

// Invoke - Ponto de entrada para chamadas do tipo Invoke.
// Funções suportadas:
// "init": inicializa o estado do chaincode, também utilizado como reset
// "registrarProposta(Id, cpfPagador, pagadorAceitou, beneficiarioAceitou, 
// boletoPago)": para registrar uma nova proposta ou atualizar uma já existente.
func (t *BoletoPropostaChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke Chaincode...")
	fmt.Println("invoke is running " + function)

	// Estrutura de Seleção para escolher qual função será executada, 
	// de acordo com a funcao chamada
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "registrarProposta" {
		return t.registrarProposta(stub, args)
	}
	fmt.Println("invoke não encontrou a func: " + function)

	return nil, errors.New("Invocação de função desconhecida: " + function)
}

// registrarProposta: função Invoke para registrar uma nova proposta, recebendo os seguintes argumentos:
// args[0]: Id. Hash que identificará a proposta
// args[1]: cpfPagador. CPF do Pagador
// args[2]: pagadorAceitou. Status de aceite do Pagador da proposta
// args[3]: beneficiarioAceitou. Status de aceite do Beneficiario da proposta
// args[4]: boletoPago. Status do Pagamento do Boleto
func (t *BoletoPropostaChaincode) registrarProposta(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("registrarProposta...")

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


	// Registra a proposta na tabela 'Proposta'
	fmt.Println("Criando Proposta Id [" + idProposta + "] para CPF nº ["+ cpfPagador +"]")
	fmt.Printf("pagadorAceitou: " + strconv.FormatBool(pagadorAceitou)) 
	fmt.Printf(" | beneficiarioAceitou: " + strconv.FormatBool(beneficiarioAceitou))
	fmt.Printf(" | boletoPago: " + strconv.FormatBool(boletoPago) + "\n")

	ok, err := stub.InsertRow(nomeTabelaProposta, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: idProposta}},
			&shim.Column{Value: &shim.Column_String_{String_: cpfPagador}},
			&shim.Column{Value: &shim.Column_Bool{Bool: pagadorAceitou}},
			&shim.Column{Value: &shim.Column_Bool{Bool: beneficiarioAceitou}},
			&shim.Column{Value: &shim.Column_Bool{Bool: boletoPago}} },
	})

	// Caso a proposta já exista
	if !ok && err == nil {
		// Trecho para atualizar uma proposta existente
		// substitui um registro existente em uma linha com o registro associado ao idProposta recebido nos argumentos
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
		fmt.Println("Proposta atualizada!")
		return nil, nil

	}

	fmt.Println("Proposta criada!")

	return nil, nil
}


// ============================================================================================================================
// Query
// ============================================================================================================================

// Query is our entry point for queries

// Query - Ponto de entrada para chamadas do tipo Query.
// Funções suportadas:
// "consultarProposta(Id)": para consultar uma proposta existente
func (t *BoletoPropostaChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Query Chaincode...")

	fmt.Println("query is running " + function)

	// Estrutura de Seleção para escolher qual função será executada, 
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
	fmt.Println("consultarProposta...")
	
	var resProposta Proposta		// Proposta
	var propostaAsBytes []byte		// retorno do json em bytes
	
	// Verifica se a quantidade de argumentos recebidas corresponde a esperada
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Obtem os valores dos argumentos e os prepara para salvar na tabela 'Proposta'
	idProposta := args[0]


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

	fmt.Println("Proposta: [%s], [%s], [%b], [%b], [%b]", resProposta.ID, resProposta.CpfPagador, resProposta.PagadorAceitou, resProposta.BeneficiarioAceitou, resProposta.BoletoPago)


	// Converter o objeto da Proposta para Bytes, para retorná-lo em formato JSON
	propostaAsBytes, err = json.Marshal(resProposta)
	if err != nil {
			return nil, fmt.Errorf("Query operation failed. Error marshaling JSON: %s", err)
	}

	// retorna o objeto em bytes
	return propostaAsBytes, nil
}