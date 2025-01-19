import { useNavigate } from '@solidjs/router'
import { onMount } from 'solid-js'
import { showToast } from '~/components/ui/toast'
import { isAuthenticated } from '~/services/auth-service'

const LoginSuccessPage = () => {
	const navigate = useNavigate()

	const checkAuth = async () => {
		if (await isAuthenticated()) {
			showToast({ title: 'Logged in successfully!', variant: 'success' })
			navigate('/login')
			return
		}

		// TODO: Handle authentication problem error
	}

	onMount(() => {
		checkAuth()
	})

	return <div></div>
}

export default LoginSuccessPage
