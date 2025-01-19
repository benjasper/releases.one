import { createContext, createSignal, ParentComponent, useContext } from 'solid-js'
import { createStore } from 'solid-js/store'
import { GetMyUserResponse, SyncResponse, ToggleUserOnboardedResponse } from '~/lib/generated/api/v1/api_pb'
import { useConnect } from './connect-context'

type State = {
	user?: GetMyUserResponse
	userLoading: () => boolean

	/**
	 * Fetches the user, if it's not already loaded. This can be called multiple times at the same time.
	 * @param needsFresh If true, will refetch the user even if it's already loaded.
	 */
	fetchUser: (needsFresh?: boolean) => Promise<GetMyUserResponse>

	sync: () => Promise<SyncResponse>

	toggleUserOnboarded: () => Promise<ToggleUserOnboardedResponse>
}

const StateContext = createContext<State>()

const StateProvider: ParentComponent = props => {
	const connect = useConnect()
	const [userPromise, setUserPromise] = createSignal<Promise<GetMyUserResponse> | undefined>(undefined)

	const [state, setState] = createStore<State>({
		userLoading: () => userPromise() !== undefined,
		fetchUser: async (needsFresh = false) => {
			if (userPromise() !== undefined) {
				return userPromise()!
			}

			if (state.user && !needsFresh) {
				return new Promise(resolve => resolve(state.user!))
			}

			const promise = connect.getMyUser({})
			setUserPromise(userPromise)

			const result = await promise

			setState('user', result)
			setUserPromise(undefined)
			return result
		},
		sync: async () => {
			return await connect.sync({})
		},
		toggleUserOnboarded: async () => {
			const result = await connect.toggleUserOnboarded({})
			await state.fetchUser(true)
			return result
		},
	})

	return <StateContext.Provider value={state}>{props.children}</StateContext.Provider>
}

const useState = () => {
	const state = useContext(StateContext)

	if (!state) {
		throw new Error('StateProvider not found')
	}

	return state
}

export { StateProvider, useState }
