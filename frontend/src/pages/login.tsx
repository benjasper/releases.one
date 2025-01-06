import { useNavigate } from '@solidjs/router'
import { AiOutlineGithub } from 'solid-icons/ai'
import { Component, onMount } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'
import { isAuthenticated } from '~/services/auth-service'

const LoginPage: Component = () => {
	const navigate = useNavigate()
	const baseUrl = import.meta.env.VITE_API_BASE_URL

	onMount(async () => {
		if (await isAuthenticated()) {
			navigate('/timeline')
		}
	})

	return (
		<a href={`${baseUrl}/api/login/github`} class={buttonVariants({ variant: 'default' })}>
			Login to GitHub <AiOutlineGithub />
		</a>
	)
}

export default LoginPage
