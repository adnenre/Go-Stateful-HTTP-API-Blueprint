package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"rest-api-blueprint/internal/errors"

	"github.com/go-playground/validator/v10"
)

//
// 🔑 Context key
//

type ctxKey string

// validatedPayloadKey is used to store validated DTO in request context
const validatedPayloadKey ctxKey = "validated_payload"

//
// 🧪 Validator instance
//

var validate = validator.New()

//
// 🧩 DTO Resolver type
//

// DTOResolver decides which DTO should be used for a request.
// Returns (dtoInstance, true) if validation should be applied.
type DTOResolver func(r *http.Request) (any, bool)

//
// 🔄 Global Validation Middleware
//

// ValidateRequest is a GLOBAL middleware (like Logging, JWT, etc.)
//
// It:
// - Applies only to body-based HTTP methods
// - Uses resolver to determine DTO
// - Validates request body
// - Stores validated DTO in context
func ValidateRequest(resolver DTOResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("VALIDATION MIDDLEWARE CALLED", "method", r.Method, "path", r.URL.Path)
			//
			// 🧠 Only validate methods that usually have a body
			//
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			default:
				next.ServeHTTP(w, r)
				return
			}

			//
			// 🧩 Resolve DTO for this request
			//
			target, ok := resolver(r)
			if !ok {
				// No DTO → skip validation
				next.ServeHTTP(w, r)
				return
			}

			//
			// 📥 Read and restore request body
			//
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
				errors.WriteProblem(w, r, errDomain, GetRequestID(r))
				return
			}

			// Restore body so controller can still read it
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Skip empty body
			if len(bodyBytes) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			//
			// 📦 Decode JSON into DTO
			//
			if err := json.Unmarshal(bodyBytes, target); err != nil {
				errDomain := errors.BadRequestError("Invalid request body: " + err.Error())
				errors.WriteProblem(w, r, errDomain, GetRequestID(r))
				return
			}

			//
			// ✅ Validate DTO
			//
			if err := validate.Struct(target); err != nil {
				fieldErrors := make(map[string]string)

				if ve, ok := err.(validator.ValidationErrors); ok {

					t := reflect.TypeOf(target)
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}

					for _, fe := range ve {
						fieldName := fe.Field()

						// Use JSON tag instead of struct field name
						if t.Kind() == reflect.Struct {
							if field, found := t.FieldByName(fe.Field()); found {
								if jsonTag := field.Tag.Get("json"); jsonTag != "" {
									fieldName = strings.Split(jsonTag, ",")[0]
								}
							}
						}

						// Basic message mapping
						switch fe.Tag() {
						case "required":
							fieldErrors[fieldName] = "This field is required"
						case "email":
							fieldErrors[fieldName] = "Must be a valid email address"
						case "min":
							fieldErrors[fieldName] = "Minimum length is " + fe.Param()
						case "oneof":
							fieldErrors[fieldName] = "Must be one of: " + fe.Param()
						default:
							fieldErrors[fieldName] = fe.Tag()
						}
					}
				}

				domainErr := errors.UnprocessableEntityError("Validation failed", fieldErrors)
				errors.WriteProblem(w, r, domainErr, GetRequestID(r))
				return
			}

			//
			// 📦 Store validated DTO in context
			//
			ctx := context.WithValue(r.Context(), validatedPayloadKey, target)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

//
// 📤 Helpers
//

// GetValidatedDTO returns raw DTO from context
func GetValidatedDTO(r *http.Request) any {
	return r.Context().Value(validatedPayloadKey)
}

// GetValidatedDTOAs returns DTO in a type-safe way
func GetValidatedDTOAs[T any](r *http.Request) (T, bool) {
	var zero T

	val := GetValidatedDTO(r)
	if val == nil {
		return zero, false
	}

	dto, ok := val.(*T)
	if !ok {
		return zero, false
	}

	return *dto, true
}
