import { makePersisted } from '@solid-primitives/storage'
import { Component, createContext, ParentComponent, useContext } from 'solid-js'
import { createStore, SetStoreFunction, Store } from 'solid-js/store'
import { RepositoryStarType } from '~/lib/generated/api/v1/api_pb'

type LocalStore = {
	settings: {
		showPrereleases: boolean
		showDescription: boolean
		selectedStarType: RepositoryStarType | null
	}
}

const Context = createContext<[get: Store<LocalStore>, set: SetStoreFunction<LocalStore>]>()

export const LocalStoreProvider: ParentComponent = props => {
	const [getLocalStorage, setLocalStorage] = makePersisted(createStore<LocalStore>({
		settings: {
			showPrereleases: true,
			showDescription: true,
			selectedStarType: null,
		},
	}), {storage: localStorage})

	return <Context.Provider value={[getLocalStorage, setLocalStorage]}>{props.children}</Context.Provider>
}

export const useLocalStore = () => {
	const store = useContext(Context)
	if (!store) throw new Error('useLocalStore must be used within a LocalStoreProvider')

	return store
}
