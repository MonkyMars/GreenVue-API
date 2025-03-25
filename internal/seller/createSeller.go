package seller

/*
	id uuid not null,
  	created_at timestamp with time zone not null default now(),
  	description text null,
  	rating integer not null default 0,
  	verified boolean not null default false,
*/

func CreateSeller(uuid int64, description string) {
	// rating and verified are set to default values. No need to provide them.
}
