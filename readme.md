[![Tests](https://github.com/dgyurics/auth/actions/workflows/tests.yaml/badge.svg)](https://github.com/dgyurics/auth/actions/workflows/tests.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/dgyurics/auth)](https://goreportcard.com/report/github.com/dgyurics/auth)

### Simple, fault-tolerant, distributed authentication service</br></br>
`api-gateway`: nginx configured with auth request module. It is the entry point for all requests.
</br>
`auth-server`: authentication server which api-gateway calls using subrequests.
</br>
`secure-server`: http server accessible to authorized users only.
</br></br>
### Instructions for running locally</br></br>
From root directory run `make run`
Create a new user `curl ...`.
Access the secure server `curl ... -H "Cookie: ..."`</br></br>

_Voilà, you've just authenticated and accessed a secure server using a distributed network of services._
