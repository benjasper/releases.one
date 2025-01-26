import { Component, createMemo, createResource, createSignal, For, onCleanup, onMount, Show } from 'solid-js'
import { useConnect } from '~/context/connect-context'
import { Card, CardContent, CardFooter, CardHeader } from '~/components/ui/card'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Button } from '~/components/ui/button'
import { Switch, SwitchControl, SwitchLabel, SwitchThumb } from '~/components/ui/switch'
import { TextField, TextFieldInput } from '~/components/ui/text-field'
import { formatDistance } from 'date-fns/formatDistance'
import { Tooltip, TooltipContent, TooltipTrigger } from '~/components/ui/tooltip'
import Navbar from '~/components/navbar'
import { Skeleton } from '~/components/ui/skeleton'
import { FiArrowUp, FiExternalLink, FiEye, FiFilter, FiStar } from 'solid-icons/fi'
import { RepositoryStarType } from '~/lib/generated/api/v1/api_pb'
import { AiFillStar } from 'solid-icons/ai'
import StarTypeSelect from '~/components/star-type-select'
import { Popover, PopoverContent, PopoverTrigger } from '~/components/ui/popover'

const TimelineSkeleton: Component = () => {
	return <For each={Array(10)}>{() => <Skeleton class="rounded-lg w-full max-w-120" height={400} />}</For>
}

const TimelinePage: Component = () => {
	const [search, setSearch] = createSignal('')
	const [descriptionEnabled, setDescriptionEnabled] = createSignal(true)
	const [prereleaseEnabled, setPrereleaseEnabled] = createSignal(true)
	const [isScrollingDown, setIsScrollingDown] = createSignal(false)
	const [starType, setStarType] = createSignal<RepositoryStarType | null>(null)
	const [now, setNow] = createSignal(Date.now())

	const connect = useConnect()
	const [timeline, { refetch: refetchTimeline }] = createResource(
		() => ({ prerelease: prereleaseEnabled(), starType: starType() ?? undefined }),
		args => connect.getRepositories(args)
	)

	const filteredTimeline = () =>
		timeline()?.timeline.filter(x => x.repositoryName.toLowerCase().includes(search().toLowerCase())) ?? []

	// Signal when scrolling down
	const handleScroll = () => {
		setIsScrollingDown(window.scrollY > 20)
	}

	// Refetch timeline after it's one minute old, when the user is coming back to the page
	const lastRefresh = new Date()
	const refetchTimelineListener = () => {
		if (lastRefresh.getTime() + 1000 * 60 < Date.now()) {
			refetchTimeline()
			lastRefresh.setTime(Date.now())
		}
	}

	onMount(() => {
		window.addEventListener('scroll', handleScroll)
		window.addEventListener('focus', () => refetchTimelineListener())
	})

	onCleanup(() => {
		window.removeEventListener('scroll', handleScroll)
		window.removeEventListener('focus', () => refetchTimelineListener())
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

					<Popover>
						<PopoverTrigger as={Button<'button'>} class="cursor-pointer">
							<FiFilter class="w-6 h-6" />
							Filters
						</PopoverTrigger>
						<PopoverContent class="flex flex-col gap-4 !w-auto">
							<Switch
								class="items-center flex gap-4 justify-between"
								checked={descriptionEnabled()}
								onChange={setDescriptionEnabled}>
								<SwitchLabel class="w-auto">Show release description</SwitchLabel>
								<SwitchControl>
									<SwitchThumb />
								</SwitchControl>
							</Switch>

							<Switch
								class="items-center flex gap-4 justify-between"
								checked={prereleaseEnabled()}
								onChange={setPrereleaseEnabled}>
								<SwitchLabel class="w-auto">Show prereleases</SwitchLabel>
								<SwitchControl>
									<SwitchThumb />
								</SwitchControl>
							</Switch>

							<StarTypeSelect starType={starType()} onChange={setStarType} />
						</PopoverContent>
					</Popover>
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
										<div class="flex items-center gap-2 justify-between">
											<a
												href={timelineItem.repositoryUrl}
												class="flex items-center no-underline hover:underline group"
												target="_blank"
												rel="noopener noreferrer">
												<span class="font-normal">{timelineItem.repositoryName}</span>
												<FiExternalLink class="opacity-0 inline-block ml-1.5 text-gray-400 w-3 transition-all group-hover:opacity-100" />
											</a>

											<Show when={timelineItem.starType === RepositoryStarType.STAR}>
												<Tooltip>
													<TooltipTrigger>
														<AiFillStar class="w-4" />
													</TooltipTrigger>
													<TooltipContent>You have starred this repository</TooltipContent>
												</Tooltip>
											</Show>

											<Show when={timelineItem.starType === RepositoryStarType.WATCH}>
												<Tooltip>
													<TooltipTrigger>
														<FiEye class="w-4" />
													</TooltipTrigger>
													<TooltipContent>You are watching this repository</TooltipContent>
												</Tooltip>
											</Show>
										</div>
										<a
											href={timelineItem.url}
											class="flex items-center mr-auto no-underline hover:underline group"
											target="_blank"
											rel="noopener noreferrer">
											<h2 class="!mt-0 !mb-0 font-normal inline-block">{timelineItem.name}</h2>
											<FiExternalLink class="opacity-0 ml-1.5 text-gray-400 w-4 transition-all group-hover:opacity-100" />
										</a>
										<Show when={descriptionEnabled()}>
											<div
												class="pt-2 prose-sm overflow-hidden break-words"
												innerHTML={timelineItem.description}></div>
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
