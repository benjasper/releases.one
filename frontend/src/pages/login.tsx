import { useNavigate } from '@solidjs/router'
import { AiOutlineGithub } from 'solid-icons/ai'
import { Component, onMount } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'
import { useState } from '~/context/state-context'
import { isAuthenticated } from '~/services/auth-service'

const LoginPage: Component = () => {
	const navigate = useNavigate()
	const state = useState()
	const baseUrl = import.meta.env.VITE_API_BASE_URL

	onMount(async () => {
		if (await isAuthenticated()) {
			const user = await state.fetchUser()

			if (!user.isOnboarded) {
				navigate('/onboarding')
				return
			}

			navigate('/timeline')
		}
	})

	return (
		<div class="flex h-dvh justify-center items-center">
			<a href={`${baseUrl}/api/login/github`} class={buttonVariants({ variant: 'default' })}>
				Login with GitHub <AiOutlineGithub />
			</a>
		</div>
	)
}

export default LoginPage
