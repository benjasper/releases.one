import { timestampDate } from '@bufbuild/protobuf/wkt'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { AuthService } from '~/lib/generated/api/v1/api_pb'

const transport = createConnectTransport({
	baseUrl: '/',
	useBinaryFormat: true,
	fetch: (url, options) => {
		return fetch(url, {credentials: 'include', ...options})
	},
})

const authService = createClient(AuthService, transport)

const ACCESS_TOKEN_EXPIRES_AT = 'access_token_expires_at'

const isAuthenticated = async (): Promise<boolean> => {
	const urlParams = new URLSearchParams(window.location.search)
	if (urlParams.get('access_token_expires_at') !== null) {
		const accessTokenExpiresAt = new Date(urlParams.get('access_token_expires_at')!)
		localStorage.setItem(ACCESS_TOKEN_EXPIRES_AT, accessTokenExpiresAt.toISOString())
		return true
	}

	if (localStorage.getItem(ACCESS_TOKEN_EXPIRES_AT) === null) {
		return false
	}

	const accessTokenExpiresAt = new Date(localStorage.getItem(ACCESS_TOKEN_EXPIRES_AT)!)

	if (accessTokenExpiresAt < new Date()) {
		return refreshToken()
	}

	return true
}

async function refreshToken(): Promise<boolean> {
	const result = await tryGetAsync(() => authService.refreshToken({}))
	if (!result.success) {
		console.error(result.error)
		return false
	}
	const response = result.result

	localStorage.setItem(ACCESS_TOKEN_EXPIRES_AT, timestampDate(response.accessTokenExpiresAt!).toISOString())

	return true
}

export function logoutAndRemoveAuthData() {
	localStorage.removeItem(ACCESS_TOKEN_EXPIRES_AT)
}

export default async function tryGetAsync<T>(getter: () => Promise<T>): Promise<
	| {
			success: true
			result: T
	  }
	| {
			success: false
			error: Error
	  }
> {
	try {
		return {
			success: true,
			result: await getter(),
		}
	} catch (error) {
		return {
			success: false,
			error: error as Error,
		}
	}
}

export { isAuthenticated }
