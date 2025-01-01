import { onMount, type Component } from 'solid-js'

import { Router, Route } from '@solidjs/router'
import LoginPage from './pages/login'
import TimelinePage from './pages/timeline'
import { ConnectProvider } from './context/connect-context'
import { ColorModeProvider, ColorModeScript, createLocalStorageManager } from '@kobalte/core'
import { MetaProvider } from '@solidjs/meta'

const App: Component = () => {
	const storageManager = createLocalStorageManager('vite-ui-theme')
	return (
		<MetaProvider>
			<ConnectProvider>
				<ColorModeScript storageType={storageManager.type} />
				<ColorModeProvider storageManager={storageManager}>
					<Router>
						<Route path="/login" component={LoginPage} />
						<Route path="/" component={TimelinePage} />
						<Route path="*" component={LoginPage} />
					</Router>
				</ColorModeProvider>
			</ConnectProvider>
		</MetaProvider>
	)
}

export default App
