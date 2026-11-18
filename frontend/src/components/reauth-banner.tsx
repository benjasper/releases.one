import { AiOutlineGithub } from 'solid-icons/ai'
import { FiAlertTriangle } from 'solid-icons/fi'
import { Component, Show } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'
import { useState } from '~/context/state-context'

const ReauthBanner: Component = () => {
	const state = useState()

	return (
		<Show when={state.user?.githubReauthRequired}>
			<div class="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between rounded-lg border border-destructive/50 p-4">
				<div class="flex gap-3 items-start">
					<FiAlertTriangle class="w-5 h-5 shrink-0 mt-0.5 text-destructive" />
					<div class="flex flex-col gap-1">
						<p class="font-medium">Your GitHub connection expired</p>
						<p class="text-sm text-muted-foreground">
							Your timeline and feeds stopped updating. Log in again to resume syncing your releases.
						</p>
					</div>
				</div>
				<a href="/api/login/github" class={buttonVariants({ variant: 'default' })}>
					Log in with GitHub <AiOutlineGithub />
				</a>
			</div>
		</Show>
	)
}

export default ReauthBanner
