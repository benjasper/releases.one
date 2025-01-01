import { AiOutlineGithub } from 'solid-icons/ai'
import { Component } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'

const LoginPage: Component = () => {
	// TODO: Make baseUrl dependent on environment
	return (
		<a href={`http://localhost/login/github`} class={buttonVariants({ variant: 'default' })}>
			Login to GitHub <AiOutlineGithub />
		</a>
	)
}

export default LoginPage
