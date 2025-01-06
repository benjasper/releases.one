import {
	Component,
	createEffect,
	createMemo,
	createResource,
	createSignal,
	For,
	onCleanup,
	onMount,
	Show,
} from 'solid-js'
import { useConnect } from '~/context/connect-context'
import { Card, CardContent, CardFooter, CardHeader } from '~/components/ui/card'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Button, buttonVariants } from '~/components/ui/button'
import { Switch, SwitchControl, SwitchDescription, SwitchLabel, SwitchThumb } from '~/components/ui/switch'
import { TextField, TextFieldInput } from '~/components/ui/text-field'
import { FiArrowUp, FiExternalLink, FiRss } from 'solid-icons/fi'
import { formatDistance } from 'date-fns/formatDistance'
import { Tooltip, TooltipContent, TooltipTrigger } from '~/components/ui/tooltip'
import Navbar from '~/components/navbar'
import { Skeleton } from '~/components/ui/skeleton'

const TimelineSkeleton: Component = () => {
	return <For each={Array(10)}>{() => <Skeleton class="rounded-lg w-full max-w-120" height={400} />}</For>
}

const TimelinePage: Component = () => {
	const connect = useConnect()
	const [timeline] = createResource(() => connect.getRepositories({}))

	const [search, setSearch] = createSignal('')
	const [descriptionEnabled, setDescriptionEnabled] = createSignal(true)
	const [isScrollingDown, setIsScrollingDown] = createSignal(false)
	const [now, setNow] = createSignal(Date.now())

	const filteredTimeline = () =>
		timeline()?.timeline.filter(x => x.repositoryName.toLowerCase().includes(search().toLowerCase())) ?? []

	// Signal when scrolling down
	const handleScroll = () => {
		setIsScrollingDown(window.scrollY > 20)
	}

	onMount(() => {
		window.addEventListener('scroll', handleScroll)
	})

	onCleanup(() => {
		window.removeEventListener('scroll', handleScroll)
	})

	setInterval(() => {
		setNow(Date.now())
	}, 1000)

	const calculateDuration = (date: Date): string => {
		return formatDistance(date, now(), { addSuffix: true })
	}

	return (
		<>
			<div class="flex flex-col gap-4 container pt-4">
				<Button
					classList={{
						'opacity-100': isScrollingDown(),
					}}
					class="fixed bottom-10 right-5 z-50 p-3 opacity-0 transition-all cursor-pointer"
					onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}>
					<FiArrowUp class="w-6 h-6" />
				</Button>
				<Navbar />
				<div class="flex gap-4 items-center justify-between md:justify-start">
					<TextField>
						<TextFieldInput
							placeholder={'Search repositories'}
							value={search()}
							onInput={e => setSearch(e.currentTarget.value)}
						/>
					</TextField>

					<Switch
						class="items-center flex gap-2"
						checked={descriptionEnabled()}
						onChange={setDescriptionEnabled}>
						<SwitchControl>
							<SwitchThumb />
						</SwitchControl>
						<SwitchLabel class="w-auto">Show release description</SwitchLabel>
					</Switch>
				</div>
				<div class="flex flex-col gap-4 items-center justify-center">
					<Show when={!timeline.loading} fallback={<TimelineSkeleton />}>
						<For each={filteredTimeline()}>
							{timelineItem => (
								<Card class="w-full max-w-120 transition-shadow duration-200">
									<CardHeader class="!p-0">
										<img
											class="rounded-t-lg aspect-2/1 object-cover"
											src={timelineItem.imageUrl}
											loading="lazy"
											alt={timelineItem.name}
										/>
									</CardHeader>
									<CardContent class="flex flex-col !pb-0 pt-4 prose dark:prose-invert">
										<a
											href={timelineItem.repositoryUrl}
											class="flex items-center no-underline hover:underline group"
											target="_blank"
											rel="noopener noreferrer">
											<span class="font-normal">{timelineItem.repositoryName}</span>
											<FiExternalLink class="opacity-0 inline-block ml-1.5 text-gray-400 w-3 transition-all group-hover:opacity-100" />
										</a>
										<a
											href={timelineItem.url}
											class="flex items-center no-underline hover:underline group"
											target="_blank"
											rel="noopener noreferrer">
											<h2 class="!mt-0 !mb-0 font-normal inline-block">{timelineItem.name}</h2>
											<FiExternalLink class="opacity-0 ml-1.5 text-gray-400 w-4 transition-all group-hover:opacity-100" />
										</a>
										<Show when={descriptionEnabled()}>
											<div class="pt-2 prose-sm" innerHTML={timelineItem.description}></div>
										</Show>
									</CardContent>
									<CardFooter class="flex justify-between text-muted-foreground !pt-2 text-sm">
										<Tooltip>
											<TooltipTrigger>
												{calculateDuration(timestampDate(timelineItem.releasedAt!))}
											</TooltipTrigger>
											<TooltipContent>
												{timestampDate(timelineItem.releasedAt!).toLocaleString()}
											</TooltipContent>
										</Tooltip>
									</CardFooter>
								</Card>
							)}
						</For>
					</Show>
				</div>
			</div>
		</>
	)
}

export default TimelinePage
