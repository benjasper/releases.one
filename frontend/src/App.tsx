import type { Component } from 'solid-js'

import logo from './logo.svg'
import { Button, buttonVariants } from './components/ui/button'
import { AiOutlineGithub } from 'solid-icons/ai'

const App: Component = () => {
	return (
		<a href={`http://localhost/login/github`} class={buttonVariants({ variant: 'default' })}>
			Login to GitHub <AiOutlineGithub />
		</a>
	)
}

export default App
