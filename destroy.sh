while getopts s: flag
do
    case "${flag}" in
        s) stack=${OPTARG};;
    esac
done

if [ -z "$stack" ]; then
        echo 'Missing -s flag' >&2
        exit 1
fi


pushd ./infrastructure &> /dev/null
    pulumi down -s $stack
popd &> /dev/null
