package adapters

/*func BindVideoLinkToSession(sessionId, filename string) error {
	client := http.Client{}
	req, _ := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/bind/%s/media?link=%s", os.Getenv("SERVERAPP_HOST"), sessionId, filename), nil)
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return errors.New("bad request: serverApp")
	}

	return nil
}
*/
