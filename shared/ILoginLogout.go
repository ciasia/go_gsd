package shared

type ILoginLogout interface {
	ForceLogin(request IRequest, email string)
	LoadUserById(id uint64) (IUser, error)
	HandleLogin(IRequest) (IResponse, error)
	HandleLogout(IRequest) (IResponse, error)
	HandleSetPassword(IRequest) (IResponse, error)

	AddAuthenticator(IAuthenticator)
	HandleOauthCallback(IRequest) (IResponse, error)
}

type IAuthenticator interface {
	TryToAuthenticate(IRequest) (string, interface{}, error)
}
