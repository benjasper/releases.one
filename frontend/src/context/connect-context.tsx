import { createContext, useContext } from 'solid-js'
import { Client, createClient, Interceptor } from '@connectrpc/connect'
import { ApiService } from '~/lib/generated/api/v1/api_pb'
import { createConnectTransport } from '@connectrpc/connect-web'
import { ParentComponent } from 'solid-js/types/server/rendering.js'
import { isAuthenticated } from '~/services/auth-service'

const abortController = new AbortController()

const authInterceptor: Interceptor = next => async req => {
	if (!(await isAuthenticated())) {
		abortController.abort()
		// We don't always have access to solid router, so we'll just redirect to login
		window.location.href = '/login'
		return await next(req)
	}

	return await next(req)
}

const baseUrl = import.meta.env.VITE_API_BASE_URL

const transport = createConnectTransport({
	baseUrl: baseUrl,
	useBinaryFormat: true,
	interceptors: [authInterceptor],
	fetch: (url, options) => {
		return fetch(url, { credentials: 'include', ...options, signal: abortController.signal })
	},
})

const ConnectContext = createContext<Client<typeof ApiService>>()

const ConnectProvider: ParentComponent = props => {
	const client = createClient(ApiService, transport)
	return <ConnectContext.Provider value={client}>{props.children}</ConnectContext.Provider>
}

const useConnect = () => {
	const client = useContext(ConnectContext)

	if (!client) {
		throw new Error('ConnectProvider not found')
	}

	return client
}

export { ConnectProvider, useConnect, transport }
