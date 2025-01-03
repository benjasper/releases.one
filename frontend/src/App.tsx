import { onMount, type Component } from 'solid-js'

import { Router, Route } from '@solidjs/router'
import LoginPage from './pages/login'
import TimelinePage from './pages/timeline'
import { ConnectProvider } from './context/connect-context'
import { ColorModeProvider, ColorModeScript, createLocalStorageManager } from '@kobalte/core'
import { MetaProvider } from '@solidjs/meta'
import LoginSuccessPage from './pages/login-success'
import { Toaster } from './components/ui/toast'

const App: Component = () => {
	const storageManager = createLocalStorageManager('vite-ui-theme')
	return (
		<MetaProvider>
			<ConnectProvider>
				<ColorModeScript storageType={storageManager.type} />
				<ColorModeProvider storageManager={storageManager}>
					<Router explicitLinks={true}>
						<Route path="/login/success" component={LoginSuccessPage} />
						<Route path="/login" component={LoginPage} />
						<Route path="/" component={TimelinePage} />
						<Route path="*" component={LoginPage} />
					</Router>
					<Toaster />
				</ColorModeProvider>
			</ConnectProvider>
		</MetaProvider>
	)
}

export default App
