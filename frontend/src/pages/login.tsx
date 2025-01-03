import { AiOutlineGithub } from 'solid-icons/ai'
import { Component } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'

const LoginPage: Component = () => {
	const baseUrl = import.meta.env.VITE_API_BASE_URL
	return (
		<a href={`${baseUrl}/api/login/github`} class={buttonVariants({ variant: 'default' })}>
			Login to GitHub <AiOutlineGithub />
		</a>
	)
}

export default LoginPage
