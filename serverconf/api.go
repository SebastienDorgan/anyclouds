package serverconf

import "io"

//UserGroup defines a user group
type UserGroup struct {
	Name  string
	Users []string
}

//User defines a user
type User struct {
	//The user's login name
	Name string
	//The user name's real name, i.e. "Bob B. Smith"
	Gecos string
	//Optional. Set to the local path you want to use. Defaults to /home/<Name>
	Homedir string
	//Optional. define the primary group. Defaults to a new group created named after the user.
	PrimaryGroup string
	//Optional. Additional groups to add the user to.
	Groups []string
	//Optional. The SELinux user for the user's login. When this is omitted the system will select the default SELinux user.
	SELinuxUser string
	//Optional. Lock the password to disable password login. Defaults to true
	LockPassword bool
	//Optional. Create the user as inactive. Defaults to false
	Inactive bool
	// The hash -- not the password itself -- of the password you want
	//            to use for this user. You can generate a safe hash via:
	//               mkpasswd --method=SHA-512 --rounds=4096
	//           (the above command would create from stdin an SHA-512 password hash
	//           with 4096 salt rounds)
	//
	//           Please note: while the use of a hashed password is better than
	//               plain text, the use of this feature is not ideal. Also,
	//               using a high number of salting rounds will help, but it should
	//               not be relied upon.
	//
	//               To highlight this risk, running John the Ripper against the
	//               example hash above, with a readily available wordlist, revealed
	//               the true password in 12 seconds on a i7-2620QM.
	//
	//               In other words, this feature is a potential security risk and is
	//               provided for your convenience only. If you do not fully trust the
	//               medium over which your cloud-config will be transmitted, then you
	//               should use SSH authentication only.
	//
	//               You have thus been warned.
	Passwd string
	//Optional. When set to true, do not create home directory. Default false
	NoCreateHome bool
	//Optional. When set to true, do not create a group named after the user. Default false
	NoUserGroup bool
	//Optional. When set to true, do not initialize lastlog and faillog database. Default false
	NoLogInit bool
	//Optional. Import SSH ids
	SSHImportIds []string
	// Optional. [list] Add keys to user's authorized keys file
	SSHAuthorizedKeys []string
	//Optional.  Set true to block ssh logins for cloud
	//     ssh public keys and emit a message redirecting logins to
	//     use <default_username> instead. This option only disables cloud
	//     provided public-keys. An error will be raised if ssh_authorized_keys
	//     or ssh_import_id is provided for the same user.
	SSHRedirectUser bool
	//Optional. Defaults to none. Accepts a sudo rule string, a list of sudo rule
	//       strings or False to explicitly deny sudo usage. Examples:
	//
	//       Allow a user unrestricted sudo access.
	//           sudo:  ALL=(ALL) NOPASSWD:ALL
	//
	//       Adding multiple sudo rule strings.
	//           sudo:
	//             - ALL=(ALL) NOPASSWD:/bin/mysql
	//             - ALL=(ALL) ALL
	//
	//       Prevent sudo access for a user.
	//           sudo: False
	//
	//       Note: Please double check your syntax and make sure it is valid.
	//             cloud-init does not parse/check the syntax of the sudo
	//             directive.
	Sudo string
	//Optional.  Create the user as a system user. This means no home directory. Default to false
	System bool
}

//ResolvConf to automatically configure resolv.conf when the instance boots for the first time.
type ResolvConf struct {
	//List of name servers. For example ["8.8.8.8", "8.8.4.4"]
	NameServers []string
	//List of search domain. For example ["foo.example.com", "bar.example.com"]
	SearchDomain []string
	//Domain For example "example.com"
	Domain string
	//Options
	Options map[string]interface{}
}

//SubnetRoute defines a subnet route
type SubnetRoute struct {
	Gateway string
	Netmask string
	Network string
}

//Subnets defines a subnet
type Subnets struct {
	DHCP         bool
	Address      string
	Netmask      string
	Gateway      string
	NameServers  []string
	SearchDomain []string
	Routes       []SubnetRoute
}

//PhysicalInterface defines a physical interface
type PhysicalInterface struct {
	Name       string
	MacAddress string
}

//NetworkRoute defines a network route
type NetworkRoute struct {
	Destination string
	Gateway     string
	Metric      *int
}

//ServerConfiguration offers a way of creating a server bootstraping script
type ServerConfiguration struct {
	Hostname           string
	DefaultUser        bool
	PackageUpgrade     bool
	Packages           []string
	Groups             []UserGroup
	Users              []User
	ResolvConf         *ResolvConf
	PhysicalInterfaces []PhysicalInterface
	NetworkRoutes      []NetworkRoute
	Extras             map[string]interface{}
}

//ConfigurationFactory abstract the creation of a server init script (cloud-init, shell, ...)
type ConfigurationFactory interface {
	Build(cfg *ServerConfiguration) (io.ByteReader, error)
}
