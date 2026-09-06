package gbcarkhos

import (
	"context"
	"log/slog"

	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/boot"
	"goark.dev/gbc-arkhos/internal/hertzlog"
	gbclog "goark.dev/gbc-log"
	"goark.dev/goark"
	goarkcontainer "goark.dev/goark/container"
	appcontext "goark.dev/goark/context"
)

// AutoConfigure 创建 Arkhos 嵌入式容器自动配置。
func AutoConfigure(options ...Option) boot.AutoConfiguration {
	copied := append([]Option(nil), options...)
	return boot.NewAutoConfiguration(
		StarterID,
		func(ctx context.Context, app *goark.ApplicationContext) error {
			if !hasConfiguration(app, gbclog.StarterID+".configuration") {
				if err := gbclog.AutoConfigure().Configure(ctx, app); err != nil {
					return err
				}
			}
			return app.RegisterConfiguration(configuration{options: copied})
		},
	)
}

type configuration struct {
	options []Option
}

func (c configuration) Name() string {
	return StarterID + ".configuration"
}

func (c configuration) Order() int {
	return 0
}

func (c configuration) Register(ctx context.Context, registry *goarkcontainer.Registry) error {
	return c.RegisterWithContext(ctx, appcontext.NewConfigurationContext(nil, registry))
}

func (c configuration) RegisterWithContext(
	_ context.Context,
	config appcontext.ConfigurationContext,
) error {
	resolved, err := newSettings(config.Environment(), c.options)
	if err != nil {
		return err
	}
	provider := resolved.resolvedProvider()
	if err := goarkcontainer.Register[*hertzlog.Bridge](
		config.Registry(),
		BeanNameHertzLogger,
		func(ctx context.Context, resolver goarkcontainer.Resolver) (*hertzlog.Bridge, error) {
			logger, err := goark.Get[*slog.Logger](ctx, resolver, gbclog.BeanNameLogger)
			if err != nil {
				return nil, err
			}
			return hertzlog.Install(logger)
		},
		goarkcontainer.WithDependsOn(gbclog.BeanNameLogger),
	); err != nil {
		return err
	}
	if err := goarkcontainer.Register[servletcontainer.Container](
		config.Registry(),
		BeanNameContainer,
		func(context.Context, goarkcontainer.Resolver) (servletcontainer.Container, error) {
			return provider.NewContainer(resolved.containerConfiguration())
		},
	); err != nil {
		return err
	}
	return goarkcontainer.Register[*EmbeddedServer](
		config.Registry(),
		BeanNameServer,
		func(ctx context.Context, resolver goarkcontainer.Resolver) (*EmbeddedServer, error) {
			container, err := goark.Get[servletcontainer.Container](
				ctx,
				resolver,
				BeanNameContainer,
			)
			if err != nil {
				return nil, err
			}
			deployments, err := goark.GetAllByType[*servletcontainer.Deployment](ctx, resolver)
			if err != nil {
				return nil, err
			}
			return NewEmbeddedServer(
				container,
				deployments,
				provider,
				resolved.serverConfiguration(),
			)
		},
		goarkcontainer.WithDependencies(BeanNameContainer),
		goarkcontainer.WithDependsOn(BeanNameHertzLogger),
	)
}

func hasConfiguration(app *appcontext.ApplicationContext, name string) bool {
	for _, descriptor := range app.Configurations() {
		if descriptor.Name == name {
			return true
		}
	}
	return false
}
