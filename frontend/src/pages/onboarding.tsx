import { A, useNavigate } from '@solidjs/router'
import { FiAlertCircle, FiCheck } from 'solid-icons/fi'
import { OcSync3 } from 'solid-icons/oc'
import { Component, createResource, createSignal, For, Match, onMount, Show } from 'solid-js'
import { Switch as SolidSwitch } from 'solid-js'
import { ParentComponent } from 'solid-js/types/server/rendering.js'
import FeedConfigurator from '~/components/feed-configurator'
import { Button, buttonVariants } from '~/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '~/components/ui/card'
import { Switch, SwitchControl, SwitchThumb } from '~/components/ui/switch'
import { useState } from '~/context/state-context'

const OnboardingPage: ParentComponent = props => {
	return (
		<div class="flex w-full h-dvh justify-center items-center">
			<Card class="w-full max-w-[380px]">{props.children}</Card>
		</div>
	)
}

const OnboardingStepOne: Component = () => {
	const state = useState()
	const [startSync, setStartSync] = createSignal(false)

	const [releases, setReleases] = createResource(startSync, () => state.sync())

	onMount(async () => {
		await state.fetchUser()
		setStartSync(true)
	})

	return (
		<>
			<CardHeader>
				<CardTitle>Welcome to releases.one</CardTitle>
				<CardDescription>
					We are syncing your starred GitHub repositories. This may take a few seconds.
				</CardDescription>
			</CardHeader>
			<CardContent class="grid gap-4">
				<SolidSwitch>
					<Match when={releases.state === 'ready'}>
						<div class="flex gap-2 justify-center items-center">
							<FiCheck class="w-4 h-4 justify-self-center" />
							<span class="text-sm">Found {releases()?.repositoryCount} repositories</span>
						</div>
					</Match>

					<Match when={releases.state === 'errored'}>
						<div class="flex gap-2 justify-center items-center">
							<FiAlertCircle class="w-4 h-4 justify-self-center" />
							<span class="text-sm">Failed to fetch repositories</span>
						</div>
							<a
								class={buttonVariants({ variant: 'default' })}
								href="https://github.com/benjasper/releases.one/issues/new"
								target="_blank">
								Report issue
							</a>
					</Match>

					<Match when={true}>
						<OcSync3 class="animate-spin w-8 h-8 justify-self-center" />
					</Match>
				</SolidSwitch>
			</CardContent>
			<Show when={releases.state === 'ready'}>
				<CardFooter class="flex gap-4 justify-center">
					<A class={buttonVariants({ variant: 'default' })} href="2">
						Continue
					</A>
				</CardFooter>
			</Show>
		</>
	)
}

const OnboardingStepTwo: Component = () => {
	const state = useState()
	const navigate = useNavigate()
	const [buttonDisabled, setButtonDisabled] = createSignal(false)

	const finishOnboarding = async () => {
		setButtonDisabled(true)
		await state.toggleUserOnboarded()
		navigate('/timeline')
	}

	return (
		<>
			<CardHeader>
				<CardTitle>Generate your feed</CardTitle>
			</CardHeader>
			<CardContent class="grid gap-4">
				<FeedConfigurator />
			</CardContent>
			<CardFooter class="flex gap-4 justify-center">
				<Button class='cursor-pointer' disabled={buttonDisabled()} onClick={finishOnboarding}>
					Continue
				</Button>
			</CardFooter>
		</>
	)
}

export { OnboardingPage, OnboardingStepOne, OnboardingStepTwo }
