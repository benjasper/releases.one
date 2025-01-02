import { createContext, useContext } from 'solid-js'
import { createClient, Interceptor } from '@connectrpc/connect'
import { ApiService } from '~/lib/generated/api/v1/api_pb'
import { createConnectTransport } from '@connectrpc/connect-web'
import { ParentComponent } from 'solid-js/types/server/rendering.js'
import { isAuthenticated } from '~/services/auth-service'
import { useNavigate } from '@solidjs/router'

const authInterceptor: Interceptor = next => async req => {
	if (!(await isAuthenticated())) {
		// const navigate = useNavigate()
		// navigate('/login')
		return await next(req)
	}

	return await next(req)
}

const transport = createConnectTransport({
	baseUrl: 'http://localhost',
	useBinaryFormat: true,
	interceptors: [authInterceptor],
	fetch: (url, options) => {
		return fetch(url, { credentials: 'include', ...options })
	},
})

const client = createClient(ApiService, transport)

const ConnectContext = createContext(client)

const ConnectProvider: ParentComponent = props => {
	return <ConnectContext.Provider value={client}>{props.children}</ConnectContext.Provider>
}

const useConnect = () => useContext(ConnectContext)

export { ConnectProvider, useConnect, transport }
