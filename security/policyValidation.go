package security

// VerifyPolicy lero-lero
func VerifyPolicy(method string, tokenlessURL string) bool {
	return true
}

// 	// ABILIO SAYS: extrair essa nojeira para um struct com paredes de chumbo para evitar vazamento e contaminação de todo o cluster com esse "shenanigan"
// 	dockerPath := strings.Split(strings.Join(strings.Split(tokenlessURL, "/")[4:], "/"), "?")[:1][0]

// 	urlServer := "http://172.24.40.63:4466"
// 	fmt.Println("Testing API...")
// 	client := swagger.NewWardenApiWithBasePath(urlServer)
// 	result, _, err := client.IsSubjectAuthorized(swagger.WardenSubjectAuthorizationRequest{
// 		Action:   strings.ToUpper(method),
// 		Resource: "srn:campus:docker:region1:sandman:dockerapi/" + dockerPath,
// 		Subject:  "weberson",
// 		Context:  ladon.Context{},
// 	})
// 	if err != nil {
// 		fmt.Printf("%v", err)
// 		os.Exit(1)
// 	}
// 	fmt.Printf("Allowed: %t\n", result.Allowed)
// 	return result.Allowed
// }

// // VerifyPolicyCtx lero-lero
// func VerifyPolicyCtx(ctx iris.Context, tokenlessURL string) bool {
// 	return VerifyPolicy(ctx.Request().Method, tokenlessURL)
// }
