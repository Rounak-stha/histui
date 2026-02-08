export default function RoadmapSection() {
	const currentFeatures = [
		{
			title: 'File Coupling Analysis',
			description: 'Detect which files change together in your Git history',
			icon: '🔗'
		},
		{
			title: 'Coupling Score',
			description: 'Get a quantified measure of file dependencies and architectural cohesion',
			icon: '📊'
		},
		{
			title: 'Refactor Opportunities',
			description: 'Identify high-impact areas where refactoring will improve code organization',
			icon: '🎯'
		}
	]

	const futureFeatures = [
		{
			title: 'Web Dashboard',
			description: 'Interactive visualization and exploration of coupling data',
			stage: 'In Development'
		},
		{
			title: 'Slack Integration',
			description: 'Receive coupling analysis reports directly in Slack',
			stage: 'Planned'
		},
		{
			title: 'CI/CD Integration',
			description: 'Enforce coupling thresholds in your deployment pipeline',
			stage: 'Planned'
		},
		{
			title: 'Team Insights',
			description: 'Track coupling trends over time and across team projects',
			stage: 'Planned'
		},
		{
			title: 'Custom Rules Engine',
			description: 'Define organization-specific patterns and coupling policies',
			stage: 'Planned'
		},
		{
			title: 'IDE Extensions',
			description: 'Get coupling insights directly in VS Code and JetBrains IDEs',
			stage: 'Early Research'
		}
	]

	return (
		<section className='w-full bg-background py-20 px-4 border-b border-border'>
			<div className='max-w-6xl mx-auto'>
				<div className='mb-16'>
					<h2 className='text-5xl font-bold text-foreground mb-4'>The histui Vision</h2>
					<p className='text-xl text-muted-foreground max-w-2xl'>
						From command-line analysis to a comprehensive platform for managing code architecture at scale.
					</p>
				</div>

				<div className='grid lg:grid-cols-2 gap-16'>
					{/* Current Features */}
					<div>
						<div className='flex items-center gap-3 mb-8'>
							<div className='w-3 h-3 bg-accent rounded-full'></div>
							<h3 className='text-2xl font-bold text-foreground'>Currently Available</h3>
						</div>
						<div className='space-y-6'>
							{currentFeatures.map((feature, idx) => (
								<div
									key={idx}
									className='p-6 border border-border bg-card/50 hover:bg-card/80 transition-colors'
								>
									<div className='flex items-start gap-4'>
										<div className='text-3xl flex-shrink-0 mt-1'>{feature.icon}</div>
										<div>
											<h4 className='font-semibold text-foreground mb-2'>{feature.title}</h4>
											<p className='text-sm text-muted-foreground'>{feature.description}</p>
										</div>
									</div>
								</div>
							))}
						</div>
					</div>

					{/* Future Features */}
					<div>
						<div className='flex items-center gap-3 mb-8'>
							<div className='w-3 h-3 bg-muted rounded-full'></div>
							<h3 className='text-2xl font-bold text-foreground'>Planned Releases</h3>
						</div>
						<div className='space-y-6'>
							{futureFeatures.map((feature, idx) => (
								<div key={idx} className='p-6 border border-border/50 bg-card/25 opacity-75'>
									<div className='flex items-start justify-between gap-4'>
										<div className='flex-1'>
											<h4 className='font-semibold text-foreground mb-2'>{feature.title}</h4>
											<p className='text-sm text-muted-foreground mb-3'>{feature.description}</p>
											<span className='inline-block text-xs font-mono text-accent px-2 py-1 bg-accent/10 rounded'>
												{feature.stage}
											</span>
										</div>
									</div>
								</div>
							))}
						</div>
					</div>
				</div>
			</div>
		</section>
	)
}
