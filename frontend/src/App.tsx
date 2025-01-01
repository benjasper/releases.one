import { onMount, type Component } from 'solid-js'

import { Router, Route } from '@solidjs/router'
import LoginPage from './pages/login'
import TimelinePage from './pages/timeline'

const App: Component = () => {
	return (
		<Router>
			<Route path="/login" component={LoginPage} />
			<Route path="/" component={TimelinePage} />
			<Route path="*" component={LoginPage} />
		</Router>
	)
}

export default App
