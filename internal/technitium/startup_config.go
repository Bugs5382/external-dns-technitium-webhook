package technitium

/*
Apache License 2.0

Copyright 2026 external-dns-technitium-webhook Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

type StartupConfig struct {
	Host       string `env:"TECHNITIUM_HOST,required" envDefault:"localhost"`
	Port       int    `env:"TECHNITIUM_PORT,required" envDefault:"5380"`
	Username   string `env:"TECHNITIUM_USER"`
	Password   string `env:"TECHNITIUM_PASSWORD"`
	Token      string `env:"TECHNITIUM_TOKEN"`
	SessionTTL int    `env:"TECHNITIUM_SESSION_TTL" envDefault:"30"`
	SSLVerify  bool   `env:"TECHNITIUM_SSL_VERIFY" envDefault:"false"`
	DryRun     bool   `env:"TECHNITIUM_DRY_RUN" envDefault:"false"`
	CreatePTR  bool   `env:"TECHNITIUM_CREATE_PTR" envDefault:"false"`
	DefaultTTL int    `env:"TECHNITIUM_DEFAULT_TTL" envDefault:"300"`
	UseTTL     bool   `env:"TECHNITIUM_USE_TTL" envDefault:"true"`
	FQDNRegEx  string
	NameRegEx  string
}
