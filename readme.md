# GoDash - simple dashboard

The purpose in building this is solely for listing down any tools or related links needed for development. 
It could include Monitoring Tools, Documentations Links, etc. 
With this, we can just visit one central dashboard, to link us everywhere.

I usually utilise this to list down my monitoring tools in every environment, and my documentation page.

Hopefully this could help anyone who has the same intention as me. :)

## How To Run Using Docker

- Prepare `config.docker.yml` in the same directory, using `config.default.yml` as reference
- Build docker image using `docker build -t <tag> .`
- Run docker image using `docker run -it -p 3500:3500 <tag>`
- Now accessible at `http://localhost:3500`
- For authentication to work, we need to make it running at a certain domain. 
For example, if we want to run locally on our laptop/computer, we could just modify `/etc/hosts` as follows:
`127.0.0.1 dashboard.domain.local`
- After that, ensure the `cookie_domain` in `config.docker.yml` changed to `.domain.local`
- If everything went smooth, you can then access it at `http://dashboard.domain.local:3500` and the authentication should also be working well
