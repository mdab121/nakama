// Copyright 2017 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"encoding/json"

	"fmt"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func RuntimeBeforeHook(runtime *Runtime, jsonpbMarshaler *jsonpb.Marshaler, jsonpbUnmarshaler *jsonpb.Unmarshaler, messageType string, envelope *Envelope, session *session) (*Envelope, error) {
	fn := runtime.GetRuntimeCallback(BEFORE, messageType)
	if fn == nil {
		return envelope, nil
	}

	strEnvelope, err := jsonpbMarshaler.MarshalToString(envelope)
	if err != nil {
		return nil, err
	}

	var jsonEnvelope map[string]interface{}
	if err = json.Unmarshal([]byte(strEnvelope), &jsonEnvelope); err != nil {
		return nil, err
	}

	userId := uuid.Nil
	handle := ""
	expiry := int64(0)
	if session != nil {
		userId = session.userID
		handle = session.handle.Load()
		expiry = session.expiry
	}

	result, fnErr := runtime.InvokeFunctionBefore(fn, userId, handle, expiry, jsonEnvelope)
	if fnErr != nil {
		return nil, fnErr
	}

	bytesEnvelope, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	resultEnvelope := &Envelope{}
	if err = jsonpbUnmarshaler.Unmarshal(bytes.NewReader(bytesEnvelope), resultEnvelope); err != nil {
		return nil, err
	}

	return resultEnvelope, nil
}

func RuntimeAfterHook(logger *zap.Logger, runtime *Runtime, jsonpbMarshaler *jsonpb.Marshaler, messageType string, envelope *Envelope, session *session) {
	fn := runtime.GetRuntimeCallback(AFTER, messageType)
	if fn == nil {
		return
	}

	strEnvelope, err := jsonpbMarshaler.MarshalToString(envelope)
	if err != nil {
		logger.Error("Failed to convert proto message to protoJSON in After invocation", zap.String("message", messageType), zap.Error(err))
		return
	}

	var jsonEnvelope map[string]interface{}
	if err = json.Unmarshal([]byte(strEnvelope), &jsonEnvelope); err != nil {
		logger.Error("Failed to convert protoJSON message to Map in After invocation", zap.String("message", messageType), zap.Error(err))
		return
	}

	userId := uuid.Nil
	handle := ""
	expiry := int64(0)
	if session != nil {
		userId = session.userID
		handle = session.handle.Load()
		expiry = session.expiry
	}

	if fnErr := runtime.InvokeFunctionAfter(fn, userId, handle, expiry, jsonEnvelope); fnErr != nil {
		logger.Error("Runtime after function caused an error", zap.String("message", messageType), zap.Error(fnErr))
	}
}

func RuntimeBeforeHookAuthentication(runtime *Runtime, jsonpbMarshaler *jsonpb.Marshaler, jsonpbUnmarshaler *jsonpb.Unmarshaler, envelope *AuthenticateRequest) (*AuthenticateRequest, error) {
	messageType := strings.TrimPrefix(fmt.Sprintf("%T", envelope.Id), "*server.")
	messageType = strings.TrimSuffix(messageType, "_")
	fn := runtime.GetRuntimeCallback(BEFORE, messageType)
	if fn == nil {
		return envelope, nil
	}

	strEnvelope, err := jsonpbMarshaler.MarshalToString(envelope)
	if err != nil {
		return nil, err
	}

	var jsonEnvelope map[string]interface{}
	if err = json.Unmarshal([]byte(strEnvelope), &jsonEnvelope); err != nil {
		return nil, err
	}

	userId := uuid.Nil
	handle := ""
	expiry := int64(0)

	result, fnErr := runtime.InvokeFunctionBefore(fn, userId, handle, expiry, jsonEnvelope)
	if fnErr != nil {
		return nil, fnErr
	}

	bytesEnvelope, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	authenticationResult := &AuthenticateRequest{}
	if err = jsonpbUnmarshaler.Unmarshal(bytes.NewReader(bytesEnvelope), authenticationResult); err != nil {
		return nil, err
	}

	return authenticationResult, nil
}

func RuntimeAfterHookAuthentication(logger *zap.Logger, runtime *Runtime, jsonpbMarshaler *jsonpb.Marshaler, envelope *AuthenticateRequest, userId uuid.UUID, handle string, expiry int64) {
	messageType := strings.TrimPrefix(fmt.Sprintf("%T", envelope.Id), "*server")
	fn := runtime.GetRuntimeCallback(AFTER, messageType)
	if fn == nil {
		return
	}

	strEnvelope, err := jsonpbMarshaler.MarshalToString(envelope)
	if err != nil {
		logger.Error("Failed to convert proto message to protoJSON in After invocation", zap.String("message", messageType), zap.Error(err))
		return
	}

	var jsonEnvelope map[string]interface{}
	if err = json.Unmarshal([]byte(strEnvelope), &jsonEnvelope); err != nil {
		logger.Error("Failed to convert protoJSON message to Map in After invocation", zap.String("message", messageType), zap.Error(err))
		return
	}

	if fnErr := runtime.InvokeFunctionAfter(fn, userId, handle, expiry, jsonEnvelope); fnErr != nil {
		logger.Error("Runtime after function caused an error", zap.String("message", messageType), zap.Error(fnErr))
	}
}
